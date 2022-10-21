package kumoi

import (
	"sync"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type Omega struct {
	agent                   base.Agent
	onMessageHandlers       sync.Map
	session                 *session
	OnMessageHandler        func(msg messages.TransitFrame)
	OnSessionMessageHandler func(msg *messages.SessionMessage)
	OnBroadcastHandler      func(msg *messages.Broadcast)
	OnClosedHandler         func()
	OnErrorHandler          func(err error)
}

func NewOmega(agent base.Agent) *Omega {
	return (&Omega{}).initWithAgent(agent)
}

func (o *Omega) initWithAgent(agent base.Agent) *Omega {
	if agent == nil {
		return nil
	}

	o.agent = agent
	if o.OnMessageHandler == nil {
		o.OnMessageHandler = func(msg messages.TransitFrame) {
		}
	}

	if o.OnSessionMessageHandler == nil {
		o.OnSessionMessageHandler = func(msg *messages.SessionMessage) {
		}
	}

	if o.OnBroadcastHandler == nil {
		o.OnBroadcastHandler = func(msg *messages.Broadcast) {
		}
	}

	if o.OnClosedHandler == nil {
		o.OnClosedHandler = func() {
		}
	}

	if o.OnErrorHandler == nil {
		o.OnErrorHandler = func(err error) {
		}
	}

	omg := o
	omg.agent.OnMessage(omg.invokeOnMessage)
	omg.agent.OnSessionMessage(omg.invokeOnSessionMessage)
	omg.agent.OnBroadcast(omg.invokeOnBroadcast)
	omg.agent.OnClosed(omg.invokeOnClosed)
	omg.agent.OnError(omg.invokeOnError)
	o.session = &session{remoteSession: remoteSession[base.Session]{session: o.agent.Session()}}
	return omg
}

func (o *Omega) Agent() base.Agent {
	return o.agent
}

func (o *Omega) Session() Session {
	return o.session
}

func (o *Omega) GetRemoteSession(sessionId string) RemoteSession[base.RemoteSession] {
	if session := o.agent.GetRemoteSession(sessionId).Get(); session != nil {
		return &remoteSession[base.RemoteSession]{session: session}
	}

	return nil
}

func (o *Omega) Ping() bool {
	return o.agent.Ping().Await().IsSuccess()
}

func (o *Omega) Broadcast(msg string) bool {
	return o.agent.Broadcast(msg).Await().IsSuccess()
}

func (o *Omega) Hello() SendFuture[*messages.Hello] {
	return wrapSendFuture[*messages.Hello](o.agent.Hello())
}

func (o *Omega) ServerTime() SendFuture[*messages.ServerTime] {
	return wrapSendFuture[*messages.ServerTime](o.agent.ServerTime())
}

func (o *Omega) Channel(channelId string) *ChannelInfo {
	if v := o.agent.GetChannelMetadata(channelId).Get(); v != nil {
		meta := value.Cast[*omega.TransitFrame](v).GetGetChannelMeta()
		channelInfo := &ChannelInfo{
			meta:  meta,
			omega: o,
		}

		return channelInfo
	}

	return nil
}

func (o *Omega) Vote(voteId string) *VoteInfo {
	if v := o.agent.GetVoteMetadata(voteId).Get(); v != nil {
		meta := value.Cast[*omega.TransitFrame](v).GetGetVoteMeta()
		voteInfo := &VoteInfo{
			meta:  meta,
			omega: o,
		}

		return voteInfo
	}

	return nil
}

func (o *Omega) PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume) Player {
	omg := o
	cp := &channelPlayer{
		omega:           omg,
		channelId:       channelId,
		targetTimestamp: targetTimestamp,
		inverse:         inverse,
		volume:          volume,
		loadFutureFunc: func(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) base.SendFuture {
			return omg.agent.PlaybackChannelMessage(channelId, targetTimestamp, inverse, volume, nextId)
		},
	}

	return cp
}

func (o *Omega) Close() concurrent.Future {
	return o.agent.Close()
}

// IsClosed
// omega is closed or not
func (o *Omega) IsClosed() bool {
	return o.agent.IsClosed()
}

func (o *Omega) invokeOnMessage(tf *omega.TransitFrame) {
	o.onMessageHandlers.Range(func(key, value any) bool {
		defer kkpanic.Log()
		if c, ok := value.(func(tf *omega.TransitFrame)); ok {
			c(tf)
		}

		return true
	})

	if ctf := messages.WrapTransitFrame(tf); ctf != nil {
		o.OnMessageHandler(ctf)
	}
}

func (o *Omega) invokeOnSessionMessage(tf *omega.TransitFrame) {
	o.OnSessionMessageHandler(value.Cast[*messages.SessionMessage](messages.WrapTransitFrame(tf)))
}

func (o *Omega) invokeOnBroadcast(tf *omega.TransitFrame) {
	o.OnBroadcastHandler(value.Cast[*messages.Broadcast](messages.WrapTransitFrame(tf)))
}

func (o *Omega) invokeOnClosed() {
	o.OnClosedHandler()
}

func (o *Omega) invokeOnError(err error) {
	o.OnErrorHandler(err)
}

func (o *Omega) CreateChannel(createChannel apirequest.CreateChannel) *CreateChannelFuture {
	omg := o
	return &CreateChannelFuture{omegaFuture[*apiresponse.CreateChannel]{
		CastFuture: omg.agent.CreateChannel(createChannel),
		omega:      omg,
	}}
}

func (o *Omega) CreateVote(createVote apirequest.CreateVote) *CreateVoteFuture {
	omg := o
	return &CreateVoteFuture{omegaFuture[*apiresponse.CreateVote]{
		CastFuture: omg.agent.CreateVote(createVote),
		omega:      omg,
	}}
}

type OmegaBuilder struct {
	ab *base.AgentBuilder
}

func NewOmegaBuilder(conf *base.Config) *OmegaBuilder {
	if conf == nil {
		return nil
	}

	return &OmegaBuilder{
		base.NewAgentBuilder(conf),
	}
}

func (b *OmegaBuilder) Connect() concurrent.CastFuture[*Omega] {
	of := concurrent.NewCastFuture[*Omega]()
	b.ab.Connect().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			of.Completable().Complete(NewOmega(f.Get().(base.Agent)))
		} else if f.IsCancelled() {
			of.Completable().Cancel()
		} else if f.IsFail() {
			println(f.Error().Error())
			of.Completable().Fail(f.Error())
		}
	}))

	return of
}

func (b *OmegaBuilder) WithEngine(engine *base.Engine) *OmegaBuilder {
	b.ab.WithEngine(engine)
	return b
}

type Player interface {
	HasNext() bool
	Next() messages.TransitFrame
	Range(func(frame messages.TransitFrame))
}
