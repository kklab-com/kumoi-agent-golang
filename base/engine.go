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
	"github.com/kklab-com/goth-bytebuf"
	"github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

var ErrCantGetServiceResource = fmt.Errorf("can't get service resource")
var ErrConnectTimeout = fmt.Errorf("connect timeout")

type _EngineHandlerTask struct {
	websocket.DefaultHandlerTask
	session *session
}

func (h *_EngineHandlerTask) WSBinary(ctx channel.HandlerContext, message *websocket.DefaultMessage, params map[string]interface{}) {
	tf := &omega.TransitFrame{}
	if err := proto.Unmarshal(message.Encoded(), tf); err != nil {
		kklogger.ErrorJ("_EngineHandlerTask.WSBinary", err.Error())
		return
	}

	if tf.GetTransitId() == 0 {
		kklogger.ErrorJ("_EngineHandlerTask.WSBinary", fmt.Sprintf("zero transit_id, data: %s", value.JsonMarshal(message.Encoded())))
	}

	if stack := tf.GetStack(); stack != nil {
		for _, tf := range stack.GetFrames() {
			h.session.transitFramePreProcess(tf)
			h.session.invokeOnRead(tf)
		}
	} else {
		h.session.transitFramePreProcess(tf)
		h.session.invokeOnRead(tf)
	}
}

func (h *_EngineHandlerTask) WSConnected(ch channel.Channel, req *http2.Request, resp *http2.Response, params map[string]interface{}) {
	h.session.ch = ch
	routine.sessionPool.Store(h.session, h.session)
}

func (h *_EngineHandlerTask) WSDisconnected(ch channel.Channel, req *http2.Request, resp *http2.Response, params map[string]interface{}) {
	h.session.invokeOnClosedHandler()
	routine.sessionPool.Delete(h.session)
}

func (h *_EngineHandlerTask) WSErrorCaught(ctx channel.HandlerContext, req *http2.Request, resp *http2.Response, msg websocket.Message, err error) {
	h.session.invokeOnErrorHandler(err)
}

type Engine struct {
	Config *Config
	APIUri string
	WSUri  string
}

func NewEngine(conf *Config) *Engine {
	return &Engine{
		Config: conf,
	}
}

func (e *Engine) discovery() {
	req, _ := http.NewRequest("GET", e.Config.discoveryUri(), nil)
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.Config.Token))
	req.Header = header
	if resp, err := http.DefaultClient.Do(req); err != nil {
		kklogger.WarnJ("base:Engine.discovery", err.Error())
	} else {
		rm := map[string]string{}
		json.Unmarshal(buf.EmptyByteBuf().WriteReader(resp.Body).Bytes(), &rm)
		resp.Body.Close()
		if str, f := rm["node_api_uri"]; f {
			e.APIUri = str
		}

		if str, f := rm["node_bs_uri"]; f {
			e.WSUri = str
		}
	}
}

func (e *Engine) checkResource() bool {
	if e.APIUri == "" || e.WSUri == "" {
		e.discovery()
	}

	if e.APIUri == "" || e.WSUri == "" {
		return false
	}

	return true
}

func (e *Engine) connect() concurrent.CastFuture[Session] {
	if !e.checkResource() {
		return concurrent.WrapCastFuture[Session](concurrent.NewFailedFuture(ErrCantGetServiceResource))
	}

	header := http.Header{}
	header.Set("Sec-WebSocket-Protocol", fmt.Sprintf("gundam, %s", e.Config.Token))
	session := newSession(e)
	chFuture := channel.NewBootstrap().
		ChannelType(&websocket.Channel{}).
		Handler(channel.NewInitializer(func(ch channel.Channel) {
			ch.Pipeline().AddLast("HANDLER", websocket.NewInvokeHandler(&_EngineHandlerTask{session: session}, nil))
		})).
		Connect(nil, &websocket.WSCustomConnectConfig{Url: e.WSUri, Header: header})

	sessionFuture := concurrent.NewFuture()
	session.connectFuture = sessionFuture
	e.connectTimeoutWatch(sessionFuture)
	chFuture.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsFail() {
			sessionFuture.Completable().Fail(f.Error())
		} else if f.IsCancelled() {
			sessionFuture.Completable().Cancel()
		}
	}))

	return concurrent.WrapCastFuture[Session](sessionFuture)
}

func (e *Engine) connectTimeoutWatch(f concurrent.Future) {
	go func() {
		<-time.NewTimer(e.Config.ConnectTimeout).C
		f.Completable().Fail(ErrConnectTimeout)
	}()
}

/*
return future object will be *apiresponse.CreateChannel
*/
func (e *Engine) createChannel(createChannel apirequest.CreateChannel) concurrent.CastFuture[*apiresponse.CreateChannel] {
	if !e.checkResource() {
		return concurrent.WrapCastFuture[*apiresponse.CreateChannel](concurrent.NewFailedFuture(ErrCantGetServiceResource))
	}

	if createChannel.IdleTimeoutSecond == 0 {
		createChannel.IdleTimeoutSecond = 60
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/d/channels", e.APIUri), buf.NewByteBufString(value.JsonMarshal(createChannel)))
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.Config.Token))
	req.Header = header
	future := concurrent.NewFuture()
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

	return concurrent.WrapCastFuture[*apiresponse.CreateChannel](future)
}

/*
return future object will be *apiresponse.CreateChannel
*/
func (e *Engine) createVote(createVote apirequest.CreateVote) concurrent.CastFuture[*apiresponse.CreateVote] {
	if !e.checkResource() {
		return concurrent.WrapCastFuture[*apiresponse.CreateVote](concurrent.NewFailedFuture(ErrCantGetServiceResource))
	}

	if createVote.IdleTimeoutSecond == 0 {
		createVote.IdleTimeoutSecond = 60
	}

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/d/votes", e.APIUri), buf.NewByteBufString(value.JsonMarshal(createVote)))
	header := http.Header{}
	header.Set(httpheadername.Authorization, fmt.Sprintf("Bearer %s", e.Config.Token))
	req.Header = header
	future := concurrent.NewFuture()
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

	return concurrent.WrapCastFuture[*apiresponse.CreateVote](future)
}
