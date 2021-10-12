package base

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kklab-com/gone-core/channel"
	http2 "github.com/kklab-com/gone-http/http"
	httpheadername "github.com/kklab-com/gone-httpheadername"
	httpstatus "github.com/kklab-com/gone-httpstatus"
	websocket "github.com/kklab-com/gone-websocket"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/buf"
	"github.com/kklab-com/goth-kkutil/concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

var ErrCantGetServiceResource = fmt.Errorf("can't get service resource")
var ErrCantFinishConnection = fmt.Errorf("can't finish connection")
var ErrConnectTimeout = fmt.Errorf("connect timeout")

const DefaultConnectTimeout = 10 * time.Second

type _EngineHandlerTask struct {
	websocket.DefaultHandlerTask
	session *Session
}

func (h *_EngineHandlerTask) WSBinary(ctx channel.HandlerContext, message *websocket.DefaultMessage, params map[string]interface{}) {
	tf := &omega.TransitFrame{}
	if err := proto.Unmarshal(message.Encoded(), tf); err != nil {
		kklogger.ErrorJ("_EngineHandlerTask.WSBinary", err.Error())
		return
	}

	if tf.GetTransitId() == 0 {
		kklogger.ErrorJ("_EngineHandlerTask.WSBinary", "zero transit_id")
	}

	if stack := tf.GetStack(); stack != nil {
		for _, tf := range stack.GetFrames() {
			h._TransitFrameProcess(tf)
		}

		return
	}

	h._TransitFrameProcess(tf)
}

func (h *_EngineHandlerTask) _TransitFrameProcess(tf *omega.TransitFrame) {
	// hello after connection established
	if hello := tf.GetHello(); hello != nil && tf.GetClass() == omega.TransitFrame_ClassRequest {
		h.session.subject = hello.GetSubject()
		h.session.name = hello.GetSubjectName()
		h.session.id = hello.GetSessionId()
		h.session.connectFuture.Completable().Complete(h.session)
	}

	// ping, auto pong reply
	if ping := tf.GetPing(); ping != nil && tf.GetClass() == omega.TransitFrame_ClassRequest {
		h.session.Send(tf.Clone().AsResponse().RenewTimestamp().SetData(&omega.TransitFrame_Pong{Pong: &omega.Pong{}}))
	}

	h.session.invokeOnRead(tf)
}

func (h *_EngineHandlerTask) WSConnected(ch channel.Channel, req *http2.Request, resp *http2.Response, params map[string]interface{}) {
	h.session.ch = ch
	routine.sessionPool.Store(h.session, h.session)
}

func (h *_EngineHandlerTask) WSDisconnected(ch channel.Channel, req *http2.Request, resp *http2.Response, params map[string]interface{}) {
	for _, f := range h.session.onDisconnectedHandler {
		kkpanic.LogCatch(func() {
			f()
		})
	}

	routine.sessionPool.Delete(h.session)
}

func (h *_EngineHandlerTask) WSErrorCaught(ctx channel.HandlerContext, req *http2.Request, resp *http2.Response, msg websocket.Message, err error) {
	for _, f := range h.session.onErrorHandler {
		kkpanic.LogCatch(func() {
			f(err)
		})
	}
}

type Engine struct {
	config *Config
	apiUri string
	wsUri  string
}

func NewEngine(conf *Config) *Engine {
	return &Engine{
		config: conf,
	}
}

func (e *Engine) discovery() {
	req, _ := http.NewRequest("GET", e.config.discoveryUri(), nil)
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.config.Token))
	req.Header = header
	if resp, err := http.DefaultClient.Do(req); err != nil {
		kklogger.WarnJ("base:Engine.discovery", err.Error())
	} else {
		rm := map[string]string{}
		json.Unmarshal(buf.EmptyByteBuf().WriteReader(resp.Body).Bytes(), &rm)
		resp.Body.Close()
		if str, f := rm["node_api_uri"]; f {
			e.apiUri = str
		}

		if str, f := rm["node_bs_uri"]; f {
			e.wsUri = str
		}
	}
}

func (e *Engine) checkResource() bool {
	if e.apiUri == "" || e.wsUri == "" {
		e.discovery()
	}

	if e.apiUri == "" || e.wsUri == "" {
		return false
	}

	return true
}

func (e *Engine) connect() SessionFuture {
	if !e.checkResource() {
		return &DefaultSessionFuture{
			Future: concurrent.NewFailedFuture(ErrCantGetServiceResource),
		}
	}

	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", fmt.Sprintf("gundam, %s", e.config.Token))
	session := newSession(e)
	chFuture := channel.NewBootstrap().
		ChannelType(&websocket.Channel{}).
		Handler(channel.NewInitializer(func(ch channel.Channel) {
			ch.Pipeline().AddLast("HANDLER", websocket.NewInvokeHandler(&_EngineHandlerTask{session: session}, nil))
		})).
		Connect(nil, &websocket.WSCustomConnectConfig{Url: e.wsUri, Header: header})

	sessionFuture := concurrent.NewFuture(nil)
	session.connectFuture = sessionFuture
	e.connectTimeoutWatch(sessionFuture)
	chFuture.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsError() {
			sessionFuture.Completable().Fail(f.Error())
		} else if f.IsCancelled() {
			sessionFuture.Completable().Cancel()
		}
	}))

	return &DefaultSessionFuture{
		Future: sessionFuture,
	}
}

func (e *Engine) connectTimeoutWatch(f concurrent.Future) {
	go func() {
		<-time.NewTimer(DefaultConnectTimeout).C
		f.Completable().Fail(ErrConnectTimeout)
	}()
}

/*
return future object will be *apiresponse.CreateChannel
*/
func (e *Engine) createChannel(createChannel apirequest.CreateChannel) concurrent.Future {
	if !e.checkResource() {
		return concurrent.NewFailedFuture(ErrCantGetServiceResource)
	}

	if createChannel.IdleTimeoutSecond == 0 {
		createChannel.IdleTimeoutSecond = 60
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/d/channels", e.apiUri), buf.NewByteBufString(value.JsonMarshal(createChannel)))
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.config.Token))
	req.Header = header
	future := concurrent.NewFuture(nil)
	go func(future concurrent.Future) {
		if resp, err := http.DefaultClient.Do(req); err != nil {
			kklogger.WarnJ("base:Engine.createChannel", err.Error())
			future.Completable().Fail(err)
			return
		} else {
			bs := buf.EmptyByteBuf().WriteReader(resp.Body).Bytes()
			resp.Body.Close()
			if resp.StatusCode == httpstatus.OK {
				respCreateChannel := &apiresponse.CreateChannel{}
				json.Unmarshal(bs, respCreateChannel)
				future.Completable().Complete(respCreateChannel)
			} else {
				respError := &omega.Error{}
				json.Unmarshal(bs, respError)
				future.Completable().Fail(respError)
			}
		}
	}(future)

	return future
}

/*
return future object will be *apiresponse.CreateChannel
*/
func (e *Engine) createVote(createVote apirequest.CreateVote) concurrent.Future {
	if !e.checkResource() {
		return concurrent.NewFailedFuture(ErrCantGetServiceResource)
	}

	if createVote.IdleTimeoutSecond == 0 {
		createVote.IdleTimeoutSecond = 60
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/d/votes", e.apiUri), buf.NewByteBufString(value.JsonMarshal(createVote)))
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.config.Token))
	req.Header = header
	future := concurrent.NewFuture(nil)
	go func(future concurrent.Future) {
		if resp, err := http.DefaultClient.Do(req); err != nil {
			kklogger.WarnJ("base:Engine.createVote", err.Error())
			future.Completable().Fail(err)
			return
		} else {
			bs := buf.EmptyByteBuf().WriteReader(resp.Body).Bytes()
			resp.Body.Close()
			if resp.StatusCode == httpstatus.OK {
				respCreateVote := &apiresponse.CreateVote{}
				json.Unmarshal(bs, respCreateVote)
				future.Completable().Complete(respCreateVote)
			} else {
				respError := &omega.Error{}
				json.Unmarshal(bs, respError)
				future.Completable().Fail(respError)
			}
		}
	}(future)

	return future
}
