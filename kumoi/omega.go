package kumoi

import (
	"sync"

	"github.com/kklab-com/gone-core/channel"
	concurrent "github.com/kklab-com/goth-concurrent"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type OmegaFuture interface {
	concurrent.Future
	Omega() *Omega
}

type DefaultOmegaFuture struct {
	concurrent.Future
}

func (f *DefaultOmegaFuture) Omega() *Omega {
	if v := f.Get(); v != nil {
		return v.(*Omega)
	}

	return nil
}

type Omega struct {
	agent                   base.Agent
	opL                     sync.Mutex
	closed                  bool
	onMessageHandlers       sync.Map
	OnSessionMessageHandler func(tf *messages.SessionMessage)
	OnBroadcastHandler      func(tf *messages.Broadcast)
	OnClosedHandler         func()
}

func (o *Omega) initWithAgent(agent base.Agent) *Omega {
	if agent == nil {
		return nil
	}

	o.agent = agent
	if o.OnSessionMessageHandler == nil {
		o.OnSessionMessageHandler = func(tf *messages.SessionMessage) {
		}
	}

	if o.OnBroadcastHandler == nil {
		o.OnBroadcastHandler = func(tf *messages.Broadcast) {
		}
	}

	if o.OnClosedHandler == nil {
		o.OnClosedHandler = func() {
		}
	}

	o.agent.OnSessionMessage(o.invokeOnSessionMessage)
	o.agent.OnBroadcast(o.invokeOnBroadcast)
	o.agent.OnClosed(o.disconnectedProcess)
	o.agent.OnClosed(o.invokeOnDisconnected)
	o.agent.OnMessage(func(tf *omega.TransitFrame) {
		o.onMessageHandlers.Range(func(key, value interface{}) bool {
			defer kkpanic.Log()
			if c, ok := value.(func(tf *omega.TransitFrame)); ok {
				c(tf)
			}

			return true
		})
	})

	return o
}

func (o *Omega) Agent() base.Agent {
	return o.agent
}

func (o *Omega) GetAgentSession() AgentSession {
	return &agentSession{session: o.agent.Session()}
}

func (o *Omega) GetRemoteSession(sessionId string) RemoteSession {
	if session := o.Agent().GetRemoteSession(sessionId).Session(); session != nil {
		return &remoteSession{session: session}
	}

	return nil
}

func (o *Omega) Ping() bool {
	return o.agent.Ping().Await().IsSuccess()
}

func (o *Omega) Broadcast(msg string) bool {
	return o.agent.Broadcast(msg).Await().IsSuccess()
}

func (o *Omega) Hello() *messages.Hello {
	if v := o.agent.Hello().Await().Get(); v != nil {
		tf := v.(*omega.TransitFrame)
		r := &messages.Hello{}
		r.ParseTransitFrame(tf)
		return r
	}

	return nil
}

func (o *Omega) Time() *messages.Time {
	if v := o.agent.Time().Await().Get(); v != nil {
		tf := v.(*omega.TransitFrame)
		r := &messages.Time{}
		r.ParseTransitFrame(tf)
		return r
	}

	return nil
}

func (o *Omega) GetChannel(channelId string) *ChannelInfo {
	if v := o.agent.GetChannelMetadata(channelId).Get(); v != nil {
		meta := castTransitFrame(v).GetGetChannelMeta()
		channelInfo := &ChannelInfo{
			channelId: meta.GetChannelId(),
			name:      meta.GetName(),
			metadata:  meta.GetData(),
			createdAt: meta.GetCreatedAt(),
			omega:     o,
		}

		return channelInfo
	}

	return nil
}

func (o *Omega) GetVote(voteId string) *VoteInfo {
	if v := o.agent.GetVoteMetadata(voteId).Get(); v != nil {
		meta := castTransitFrame(v).GetGetVoteMeta()
		voteInfo := &VoteInfo{
			voteId:    meta.GetVoteId(),
			name:      meta.GetName(),
			metadata:  meta.GetData(),
			createdAt: meta.GetCreatedAt(),
			omega:     o,
		}

		for _, vto := range meta.GetVoteOptions() {
			voteInfo.voteOptions = append(voteInfo.voteOptions, VoteOption{
				Id:   vto.GetId(),
				Name: vto.GetName(),
			})
		}

		return voteInfo
	}

	return nil
}

func (o *Omega) PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume) TFPlayer {
	cp := &ChannelPlayer{
		omega:     o,
		channelId: channelId,
	}

	cp.load(o.Agent().PlaybackChannelMessage(channelId, targetTimestamp, inverse, volume, ""))
	if len(cp.tfs) == 0 {
		cp = nil
	}

	return cp
}

func (o *Omega) Close() concurrent.Future {
	o.closed = true
	return o.Agent().Close()
}

// IsClosed
// omega is closed or not
func (o *Omega) IsClosed() bool {
	return o.closed
}

// IsDisconnected
// session is active or not
func (o *Omega) IsDisconnected() bool {
	return !o.Agent().Session().Ch().IsActive()
}

func (o *Omega) invokeOnSessionMessage(tf *omega.TransitFrame) {
	r := &messages.SessionMessage{}
	r.ParseTransitFrame(tf)
	o.OnSessionMessageHandler(r)
}

func (o *Omega) invokeOnBroadcast(tf *omega.TransitFrame) {
	r := &messages.Broadcast{}
	r.ParseTransitFrame(tf)
	o.OnBroadcastHandler(r)
}

func (o *Omega) invokeOnDisconnected() {
	o.OnClosedHandler()
}

type CreateChannelFuture interface {
	concurrent.Future
	Response() *apiresponse.CreateChannel
	Info() *ChannelInfo
	Join() *Channel
}

type DefaultCreateChannelFuture struct {
	concurrent.Future
	omega *Omega
}

func (f *DefaultCreateChannelFuture) Response() *apiresponse.CreateChannel {
	if v := f.Get(); v != nil {
		return v.(*apiresponse.CreateChannel)
	}

	return nil
}

func (f *DefaultCreateChannelFuture) Info() *ChannelInfo {
	if resp := f.Response(); resp != nil {
		return f.omega.GetChannel(resp.ChannelId)
	}

	return nil
}

func (f *DefaultCreateChannelFuture) Join() *Channel {
	if resp := f.Response(); resp != nil {
		return f.Info().Join(resp.OwnerKey)
	}

	return nil
}

func (o *Omega) CreateChannel(createChannel apirequest.CreateChannel) CreateChannelFuture {
	cf := o.Agent().CreateChannel(createChannel)
	ccf := &DefaultCreateChannelFuture{
		Future: channel.NewFuture(nil),
		omega:  o,
	}

	cf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			ccf.Completable().Complete(f.Get())
		} else if f.IsFail() {
			ccf.Completable().Fail(f.Error())
		} else if f.IsCancelled() {
			ccf.Completable().Cancel()
		}
	}))

	return ccf
}

type CreateVoteFuture interface {
	concurrent.Future
	Response() *apiresponse.CreateVote
	Info() *VoteInfo
	Join() *Vote
}

type DefaultCreateVoteFuture struct {
	concurrent.Future
	omega *Omega
}

func (f *DefaultCreateVoteFuture) Response() *apiresponse.CreateVote {
	if v := f.Get(); v != nil {
		return v.(*apiresponse.CreateVote)
	}

	return nil
}

func (f *DefaultCreateVoteFuture) Info() *VoteInfo {
	if resp := f.Response(); resp != nil {
		return f.omega.GetVote(resp.VoteId)
	}

	return nil
}

func (f *DefaultCreateVoteFuture) Join() *Vote {
	if resp := f.Response(); resp != nil {
		return f.Info().Join(resp.Key)
	}

	return nil
}

func (o *Omega) CreateVote(createVote apirequest.CreateVote) CreateVoteFuture {
	vf := o.Agent().CreateVote(createVote)
	cvf := &DefaultCreateVoteFuture{
		Future: channel.NewFuture(nil),
		omega:  o,
	}

	vf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			cvf.Completable().Complete(f.Get())
		} else if f.IsFail() {
			cvf.Completable().Fail(f.Error())
		} else if f.IsCancelled() {
			cvf.Completable().Cancel()
		}
	}))

	return cvf
}

func (o *Omega) disconnectedProcess() {
}

type OmegaBuilder struct {
	engine *base.Engine
}

func NewOmegaBuilder(engine *base.Engine) *OmegaBuilder {
	if engine == nil {
		return nil
	}

	return &OmegaBuilder{engine: engine}
}

func (b *OmegaBuilder) Connect() OmegaFuture {
	if b.engine == nil || b.engine.Config == nil {
		return &DefaultOmegaFuture{Future: concurrent.NewFailedFuture(base.ErrConfigIsEmpty)}
	}

	of := &DefaultOmegaFuture{Future: concurrent.NewFuture()}
	base.NewAgentBuilder(b.engine).Connect().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			of.Completable().Complete((&Omega{}).initWithAgent(f.Get().(base.Agent)))
		} else if f.IsCancelled() {
			of.Completable().Cancel()
		} else if f.IsFail() {
			of.Completable().Fail(f.Error())
		}
	}))

	return of
}

func castTransitFrame(v interface{}) *omega.TransitFrame {
	if v == nil {
		return nil
	}

	if c, ok := v.(*omega.TransitFrame); ok {
		return c
	}

	return nil
}
