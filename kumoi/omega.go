package kumoi

import (
	"sync"

	"github.com/kklab-com/goth-kkutil/concurrent"
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
	sf base.AgentFuture
}

func (f *DefaultOmegaFuture) Omega() *Omega {
	if v := f.Get(); v != nil {
		return v.(*Omega)
	}

	return nil
}

type Omega struct {
	agent                   *base.Agent
	opL                     sync.Mutex
	closed                  bool
	onMessageHandlers       sync.Map
	OnSessionMessageHandler func(tf *messages.SessionMessage)
	OnBroadcastHandler      func(tf *messages.Broadcast)
	OnDisconnectedHandler   func()
}

func (o *Omega) initWithAgent(agent *base.Agent) *Omega {
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

	if o.OnDisconnectedHandler == nil {
		o.OnDisconnectedHandler = func() {
		}
	}

	o.agent.OnSessionMessage(o.invokeOnSessionMessage)
	o.agent.OnBroadcast(o.invokeOnBroadcast)
	o.agent.OnDisconnected(o.disconnectedProcess)
	o.agent.OnDisconnected(o.invokeOnDisconnected)
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

func (o *Omega) Agent() *base.Agent {
	return o.agent
}

func (o *Omega) GetAgentSession() base.AgentSession {
	return o.agent.Session()
}

func (o *Omega) GetRemoteSession(sessionId string) base.RemoteSession {
	return o.Agent().GetRemoteSession(sessionId).Session()
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
		meta := convertTransitFrame(v).GetGetChannelMeta()
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
		meta := convertTransitFrame(v).GetGetVoteMeta()
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
	return o.Agent().Disconnect()
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
	o.OnDisconnectedHandler()
}

func (o *Omega) CreateChannel(createChannel apirequest.CreateChannel) *apiresponse.CreateChannel {
	return o.Agent().CreateChannel(createChannel).Get().(*apiresponse.CreateChannel)
}

func (o *Omega) CreateVote(createVote apirequest.CreateVote) *apiresponse.CreateVote {
	return o.Agent().CreateVote(createVote).Get().(*apiresponse.CreateVote)
}

func (o *Omega) disconnectedProcess() {
}

type OmegaBuilder struct {
	config *base.Config
}

func NewOmegaBuilder(conf *base.Config) *OmegaBuilder {
	if conf == nil {
		return nil
	}

	return &OmegaBuilder{config: conf}
}

func (b *OmegaBuilder) Connect() OmegaFuture {
	if b.config == nil {
		return &DefaultOmegaFuture{Future: concurrent.NewFailedFuture(base.ErrConfigIsEmpty)}
	}

	af := base.NewAgentBuilder(b.config).Connect()
	of := &DefaultOmegaFuture{Future: concurrent.NewFuture(af.Ctx()), sf: af}
	of.sf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			of.Completable().Complete((&Omega{}).initWithAgent(of.sf.Get().(*base.Agent)))
		}
	}))

	return of
}

func convertTransitFrame(v interface{}) *omega.TransitFrame {
	if v == nil {
		return nil
	}

	if c, ok := v.(*omega.TransitFrame); ok {
		return c
	}

	return nil
}
