package kumoi

import (
	"sync"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SendFuture[T messages.TransitFrame] interface {
	Base() base.SendFuture
	Await() SendFuture[T]
	AwaitTimeout(timeout time.Duration) SendFuture[T]
	IsDone() bool
	IsSuccess() bool
	IsCancelled() bool
	IsFail() bool
	Error() error
	TransitFrame() (t T)
}

func wrapSendFuture[T messages.TransitFrame](sf base.SendFuture) (t SendFuture[T]) {
	return value.Cast[SendFuture[T]](&DefaultSendFuture[T]{bf: sf})
}

type DefaultSendFuture[T messages.TransitFrame] struct {
	bf base.SendFuture
}

func (f *DefaultSendFuture[T]) Base() base.SendFuture {
	return f.bf
}

func (f *DefaultSendFuture[T]) Await() SendFuture[T] {
	f.Base().Await()
	return f
}

func (f *DefaultSendFuture[T]) AwaitTimeout(timeout time.Duration) SendFuture[T] {
	f.Base().AwaitTimeout(timeout)
	return f
}

func (f *DefaultSendFuture[T]) IsDone() bool {
	return f.Base().IsDone()
}

func (f *DefaultSendFuture[T]) IsSuccess() bool {
	return f.Base().IsSuccess()
}

func (f *DefaultSendFuture[T]) IsCancelled() bool {
	return f.Base().IsCancelled()
}

func (f *DefaultSendFuture[T]) IsFail() bool {
	return f.Base().IsFail()
}

func (f *DefaultSendFuture[T]) Error() error {
	return f.Base().Error()
}

func (f *DefaultSendFuture[T]) TransitFrame() (t T) {
	var an any = getParsedTransitFrameFromBaseTransitFrame(f.bf.TransitFrame())
	t = value.Cast[T](an)
	return
}

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
	onMessageHandlers       sync.Map
	OnMessageHandler        func(tf messages.TransitFrame)
	OnSessionMessageHandler func(tf *messages.SessionMessage)
	OnBroadcastHandler      func(tf *messages.Broadcast)
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
		o.OnMessageHandler = func(tf messages.TransitFrame) {
		}
	}

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
	return omg
}

func (o *Omega) Agent() base.Agent {
	return o.agent
}

func (o *Omega) Session() Session {
	return &session{remoteSession: remoteSession[base.Session]{session: o.agent.Session()}}
}

func (o *Omega) GetRemoteSession(sessionId string) RemoteSession[base.RemoteSession] {
	if session := o.agent.GetRemoteSession(sessionId).Session(); session != nil {
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

func (o *Omega) Vote(voteId string) *VoteInfo {
	if v := o.agent.GetVoteMetadata(voteId).Get(); v != nil {
		meta := value.Cast[*omega.TransitFrame](v).GetGetVoteMeta()
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

	if ctf := getParsedTransitFrameFromBaseTransitFrame(tf); ctf != nil {
		o.OnMessageHandler(ctf)
	}
}

func (o *Omega) invokeOnSessionMessage(tf *omega.TransitFrame) {
	o.OnSessionMessageHandler(value.Cast[*messages.SessionMessage](getParsedTransitFrameFromBaseTransitFrame(tf)))
}

func (o *Omega) invokeOnBroadcast(tf *omega.TransitFrame) {
	o.OnBroadcastHandler(value.Cast[*messages.Broadcast](getParsedTransitFrameFromBaseTransitFrame(tf)))
}

func (o *Omega) invokeOnClosed() {
	o.OnClosedHandler()
}

func (o *Omega) invokeOnError(err error) {
	o.OnErrorHandler(err)
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
		return f.omega.Channel(resp.ChannelId)
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
	cf := o.agent.CreateChannel(createChannel)
	ccf := &DefaultCreateChannelFuture{
		Future: concurrent.NewFuture(),
		omega:  o,
	}

	cf.Base().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
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
		return f.omega.Vote(resp.VoteId)
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
	vf := o.agent.CreateVote(createVote)
	cvf := &DefaultCreateVoteFuture{
		Future: concurrent.NewFuture(),
		omega:  o,
	}

	vf.Base().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
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

type OmegaBuilder struct {
	conf *base.Config
}

func NewOmegaBuilder(conf *base.Config) *OmegaBuilder {
	if conf == nil {
		return nil
	}

	return &OmegaBuilder{conf: conf}
}

func (b *OmegaBuilder) Connect() OmegaFuture {
	if b.conf == nil {
		return &DefaultOmegaFuture{Future: concurrent.NewFailedFuture(base.ErrConfigIsEmpty)}
	}

	of := &DefaultOmegaFuture{Future: concurrent.NewFuture()}
	base.NewAgentBuilder(b.conf).Connect().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
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

type Player interface {
	Next() messages.TransitFrame
}

func getParsedTransitFrameFromBaseTransitFrame(btf *omega.TransitFrame) (tf messages.TransitFrame) {
	if btf == nil {
		return nil
	}

	switch btf.GetData().(type) {
	case *omega.TransitFrame_Broadcast:
		tf = &messages.Broadcast{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_Hello:
		tf = &messages.Hello{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_ServerTime:
		tf = &messages.ServerTime{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_JoinChannel:
		tf = &messages.JoinChannel{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_GetChannelMeta:
		tf = &messages.GetChannelMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_SetChannelMeta:
		tf = &messages.SetChannelMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_ChannelMessage:
		tf = &messages.ChannelMessage{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_ChannelOwnerMessage:
		tf = &messages.ChannelOwnerMessage{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_ChannelCount:
		tf = &messages.ChannelCount{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_LeaveChannel:
		tf = &messages.LeaveChannel{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_CloseChannel:
		tf = &messages.CloseChannel{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_JoinVote:
		tf = &messages.JoinVote{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_GetVoteMeta:
		tf = &messages.GetVoteMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_SetVoteMeta:
		tf = &messages.SetVoteMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_VoteMessage:
		tf = &messages.VoteMessage{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_VoteOwnerMessage:
		tf = &messages.VoteOwnerMessage{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_VoteCount:
		tf = &messages.VoteCount{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_VoteSelect:
		tf = &messages.VoteSelect{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_VoteStatus:
		tf = &messages.VoteStatus{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_LeaveVote:
		tf = &messages.LeaveVote{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_CloseVote:
		tf = &messages.CloseVote{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_GetSessionMeta:
		tf = &messages.GetSessionMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_SetSessionMeta:
		tf = &messages.SetSessionMeta{TransitFrame: messages.WrapTransitFrame(btf)}
	case *omega.TransitFrame_SessionMessage:
		tf = &messages.SessionMessage{TransitFrame: messages.WrapTransitFrame(btf)}
	}

	return tf
}
