package base

import (
	concurrent "github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type AgentFuture interface {
	concurrent.Future
	Agent() Agent
}

type DefaultAgentFuture struct {
	concurrent.Future
}

func (f *DefaultAgentFuture) Agent() Agent {
	if v := f.Get(); v != nil {
		return v.(*agent)
	}

	return nil
}

type Agent interface {
	Close() concurrent.Future
	IsClosed() bool
	Session() Session
	GetRemoteSession(sessionId string) RemoteSessionFuture
	OnClosed(f func())
	OnMessage(f func(tf *omega.TransitFrame))
	OnSessionMessage(f func(tf *omega.TransitFrame))
	OnRequest(f func(tf *omega.TransitFrame))
	OnResponse(f func(tf *omega.TransitFrame))
	OnRequestReplay(f func(tf *omega.TransitFrame))
	OnResponseReplay(f func(tf *omega.TransitFrame))
	OnNotification(f func(tf *omega.TransitFrame))
	OnNotificationReplay(f func(tf *omega.TransitFrame))
	OnError(f func(err error))
	Ping() concurrent.Future
	OnBroadcast(f func(tf *omega.TransitFrame))
	Broadcast(msg string) SendFuture
	Hello() SendFuture
	ServerTime() SendFuture
	PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture
	GetChannelMetadata(channelId string) SendFuture
	SetChannelMetadata(channelId string, name string, metadata *Metadata, skill *omega.Skill) SendFuture
	JoinChannel(channelId string, key string) SendFuture
	LeaveChannel(channelId string) SendFuture
	CloseChannel(channelId, key string) SendFuture
	ChannelMessage(channelId string, message string, metadata *Metadata) SendFuture
	ChannelCount(channelId string) SendFuture
	ChannelOwnerMessage(channelId string, message string, metadata *Metadata) SendFuture
	ReplayChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture
	GetVoteMetadata(voteId string) SendFuture
	SetVoteMetadata(voteId string, name string, metadata *Metadata) SendFuture
	JoinVote(voteId string, key string) SendFuture
	LeaveVote(voteId string) SendFuture
	CloseVote(voteId, key string) SendFuture
	VoteMessage(voteId string, message string) SendFuture
	VoteSelect(voteId string, voteOptionId string) SendFuture
	VoteCount(voteId string) SendFuture
	VoteOwnerMessage(voteId string, message string) SendFuture
	VoteStatus(voteId string, statusType omega.Vote_Status) SendFuture
	GetSessionMeta(sessionId string) SendFuture
	SetSessionMeta(metadata *Metadata) SendFuture
	SessionMessage(sessionId string, message string) SendFuture
	SessionsMessage(sessionIds []string, message string) SendFuture
	CreateChannel(createChannel apirequest.CreateChannel) CastFuture[*apiresponse.CreateChannel]
	CreateVote(createVote apirequest.CreateVote) CastFuture[*apiresponse.CreateVote]
}

type agent struct {
	session                     Session
	onClosedHandler             func()
	onMessageHandler            func(tf *omega.TransitFrame)
	onSessionMessageHandler     func(tf *omega.TransitFrame)
	onBroadcastHandler          func(tf *omega.TransitFrame)
	onRequestHandler            func(tf *omega.TransitFrame)
	onResponseHandler           func(tf *omega.TransitFrame)
	onRequestReplayHandler      func(tf *omega.TransitFrame)
	onResponseReplayHandler     func(tf *omega.TransitFrame)
	onNotificationHandler       func(tf *omega.TransitFrame)
	onNotificationReplayHandler func(tf *omega.TransitFrame)
	onErrorHandler              func(err error)
}

func NewAgent(session Session) Agent {
	if session == nil {
		return nil
	}

	agent := &agent{
		session: session,
		onErrorHandler: func(err error) {
			kklogger.ErrorJ("base:agent.onErrorHandler", err.Error())
		},
	}

	sfa := value.Cast[sessionForAgent](session)
	if sfa == nil {
		kklogger.ErrorJ("base:NewAgent", "not native session object")
		return nil
	}

	sfa.agentOnMessage(func(msg *omega.TransitFrame) {
		agent.invokeOnSessionMessage(msg)
	})

	sfa.agentOnRead(func(tf *omega.TransitFrame) {
		agent.invokeOnRead(tf)
	})

	sfa.agentOnClosed(func() {
		agent.invokeOnClosed()
	})

	sfa.agentOnError(func(err error) {
		agent.invokeOnError(err)
	})

	return agent
}

func (a *agent) invokeOnSessionMessage(tf *omega.TransitFrame) {
	if ah := a.onSessionMessageHandler; ah != nil {
		ah(tf)
	}
}

func (a *agent) invokeOnRead(tf *omega.TransitFrame) {
	var ah func(tf *omega.TransitFrame)
	switch tf.GetClass() {
	case omega.TransitFrame_ClassRequest:
		ah = a.onRequestHandler
	case omega.TransitFrame_ClassResponse:
		ah = a.onResponseHandler
	case omega.TransitFrame_ClassRequestReplay:
		ah = a.onRequestReplayHandler
	case omega.TransitFrame_ClassResponseReplay:
		ah = a.onResponseReplayHandler
	case omega.TransitFrame_ClassNotification:
		if tf.GetBroadcast() != nil {
			if iah := a.onBroadcastHandler; iah != nil {
				iah(tf)
			}
		}

		ah = a.onNotificationHandler
	case omega.TransitFrame_ClassNotificationReplay:
		ah = a.onNotificationReplayHandler
	}

	if ah != nil {
		ah(tf)
	}

	if ah := a.onMessageHandler; ah != nil {
		ah(tf)
	}
}

func (a *agent) invokeOnError(err error) {
	if ah := a.onErrorHandler; ah != nil {
		ah(err)
	}
}

func (a *agent) invokeOnClosed() {
	if ah := a.onClosedHandler; ah != nil {
		ah()
	}
}

func (a *agent) Close() concurrent.Future {
	return a.session.Close()
}

func (a *agent) IsClosed() bool {
	return a.session.IsClosed()
}

func (a *agent) Session() Session {
	return a.session
}

func (a *agent) GetRemoteSession(sessionId string) RemoteSessionFuture {
	return a.session.GetRemoteSession(sessionId)
}

func (a *agent) OnClosed(f func()) {
	a.onClosedHandler = f
}

func (a *agent) OnMessage(f func(tf *omega.TransitFrame)) {
	a.onMessageHandler = f
}

func (a *agent) OnSessionMessage(f func(tf *omega.TransitFrame)) {
	a.onSessionMessageHandler = f
}

func (a *agent) OnRequest(f func(tf *omega.TransitFrame)) {
	a.onRequestHandler = f
}

func (a *agent) OnResponse(f func(tf *omega.TransitFrame)) {
	a.onResponseHandler = f
}

func (a *agent) OnRequestReplay(f func(tf *omega.TransitFrame)) {
	a.onRequestReplayHandler = f
}

func (a *agent) OnResponseReplay(f func(tf *omega.TransitFrame)) {
	a.onResponseReplayHandler = f
}

func (a *agent) OnNotification(f func(tf *omega.TransitFrame)) {
	a.onNotificationHandler = f
}

func (a *agent) OnNotificationReplay(f func(tf *omega.TransitFrame)) {
	a.onNotificationReplayHandler = f
}

func (a *agent) OnError(f func(err error)) {
	a.onErrorHandler = f
}

func (a *agent) Ping() concurrent.Future {
	return a.session.Ping()
}

func (a *agent) OnBroadcast(f func(tf *omega.TransitFrame)) {
	a.onBroadcastHandler = f
}

func (a *agent) Broadcast(msg string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_Broadcast{Broadcast: &omega.Broadcast{Message: msg}})
}

func (a *agent) Hello() SendFuture {
	return a.session.Hello()
}

func (a *agent) ServerTime() SendFuture {
	return a.session.ServerTime()
}

func (a *agent) PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_PlaybackChannelMessage{PlaybackChannelMessage: &omega.PlaybackChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *agent) GetChannelMetadata(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetChannelMeta{GetChannelMeta: &omega.GetChannelMeta{ChannelId: channelId}})
}

func (a *agent) SetChannelMetadata(channelId string, name string, metadata *Metadata, skill *omega.Skill) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SetChannelMeta{SetChannelMeta: &omega.SetChannelMeta{ChannelId: channelId, Name: name, Data: metadata, Skill: skill}})
}

func (a *agent) JoinChannel(channelId string, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_JoinChannel{JoinChannel: &omega.JoinChannel{ChannelId: channelId, Key: key}})
}

func (a *agent) LeaveChannel(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_LeaveChannel{LeaveChannel: &omega.LeaveChannel{ChannelId: channelId}})
}

func (a *agent) CloseChannel(channelId, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_CloseChannel{CloseChannel: &omega.CloseChannel{ChannelId: channelId, Key: key}})
}

func (a *agent) ChannelMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelMessage{ChannelMessage: &omega.ChannelMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *agent) ChannelCount(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelCount{ChannelCount: &omega.ChannelCount{ChannelId: channelId}})
}

func (a *agent) ChannelOwnerMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelOwnerMessage{ChannelOwnerMessage: &omega.ChannelOwnerMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *agent) ReplayChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ReplayChannelMessage{ReplayChannelMessage: &omega.ReplayChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *agent) GetVoteMetadata(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetVoteMeta{GetVoteMeta: &omega.GetVoteMeta{VoteId: voteId}})
}

func (a *agent) SetVoteMetadata(voteId string, name string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SetVoteMeta{SetVoteMeta: &omega.SetVoteMeta{VoteId: voteId, Name: name, Data: metadata}})
}

func (a *agent) JoinVote(voteId string, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_JoinVote{JoinVote: &omega.JoinVote{VoteId: voteId, Key: key}})
}

func (a *agent) LeaveVote(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_LeaveVote{LeaveVote: &omega.LeaveVote{VoteId: voteId}})
}

func (a *agent) CloseVote(voteId, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_CloseVote{CloseVote: &omega.CloseVote{VoteId: voteId, Key: key}})
}

func (a *agent) VoteMessage(voteId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteMessage{VoteMessage: &omega.VoteMessage{VoteId: voteId, Message: message}})
}

func (a *agent) VoteSelect(voteId string, voteOptionId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteSelect{VoteSelect: &omega.VoteSelect{VoteId: voteId, VoteOptionId: voteOptionId}})
}

func (a *agent) VoteCount(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteCount{VoteCount: &omega.VoteCount{VoteId: voteId}})
}

func (a *agent) VoteOwnerMessage(voteId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteOwnerMessage{VoteOwnerMessage: &omega.VoteOwnerMessage{VoteId: voteId, Message: message}})
}

func (a *agent) VoteStatus(voteId string, statusType omega.Vote_Status) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteStatus{VoteStatus: &omega.VoteStatus{VoteId: voteId, Status: statusType}})
}

func (a *agent) GetSessionMeta(sessionId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetSessionMeta{GetSessionMeta: &omega.GetSessionMeta{SessionId: sessionId}})
}

func (a *agent) SetSessionMeta(metadata *Metadata) SendFuture {
	return a.Session().SetMetadata(metadata)
}

func (a *agent) SessionMessage(sessionId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{ToSession: sessionId, Message: message}})
}

func (a *agent) SessionsMessage(sessionIds []string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SessionsMessage{SessionsMessage: &omega.SessionsMessage{ToSessions: sessionIds, Message: message}})
}

func (a *agent) CreateChannel(createChannel apirequest.CreateChannel) CastFuture[*apiresponse.CreateChannel] {
	return a.Session().GetEngine().createChannel(createChannel)
}

func (a *agent) CreateVote(createVote apirequest.CreateVote) CastFuture[*apiresponse.CreateVote] {
	return a.Session().GetEngine().createVote(createVote)
}

type AgentBuilder struct {
	conf *Config
}

func NewAgentBuilder(conf *Config) *AgentBuilder {
	if conf == nil {
		return nil
	}

	return &AgentBuilder{conf: conf}
}

func (b *AgentBuilder) Connect() AgentFuture {
	if b.conf == nil {
		return &DefaultAgentFuture{Future: concurrent.NewFailedFuture(ErrConfigIsEmpty)}
	}

	af := &DefaultAgentFuture{Future: concurrent.NewFuture()}
	NewEngine(b.conf).connect().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			af.Completable().Complete(NewAgent(f.Get().(Session)))
		} else if f.IsCancelled() {
			af.Completable().Cancel()
		} else {
			af.Completable().Fail(f.Error())
		}
	}))

	return af
}
