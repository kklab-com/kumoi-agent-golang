package base

import (
	concurrent "github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
	kkpanic "github.com/kklab-com/goth-panic"
	"github.com/kklab-com/kumoi-agent-golang/base/apirequest"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type AgentFuture interface {
	concurrent.Future
	Agent() *Agent
}

type DefaultAgentFuture struct {
	concurrent.Future
	sf SessionFuture
}

func (f *DefaultAgentFuture) Agent() *Agent {
	if v := f.Get(); v != nil {
		return v.(*Agent)
	}

	return nil
}

type Agent struct {
	session                     *Session
	onDisconnectedHandler       []func()
	onMessageHandler            []func(tf *omega.TransitFrame)
	onSessionMessageHandler     []func(tf *omega.TransitFrame)
	onBroadcastHandler          []func(tf *omega.TransitFrame)
	onRequestHandler            []func(tf *omega.TransitFrame)
	onResponseHandler           []func(tf *omega.TransitFrame)
	onRequestReplayHandler      []func(tf *omega.TransitFrame)
	onResponseReplayHandler     []func(tf *omega.TransitFrame)
	onNotificationHandler       []func(tf *omega.TransitFrame)
	onNotificationReplayHandler []func(tf *omega.TransitFrame)
	onErrorHandler              []func(err error)
}

func NewAgent(session *Session) *Agent {
	if session == nil {
		return nil
	}

	agent := &Agent{
		session: session,
		onErrorHandler: []func(err error){func(err error) {
			kklogger.ErrorJ("base:Agent.onErrorHandler", err.Error())
		}},
	}

	session.OnMessage(func(msg *omega.TransitFrame) {
		agent.InvokeOnSessionMessage(msg)
	})

	session.OnRead(func(tf *omega.TransitFrame) {
		agent.InvokeOnRead(tf)
	})

	session.OnDisconnected(func() {
		agent.InvokeOnDisconnected()
	})

	session.OnError(func(err error) {
		agent.InvokeOnError(err)
	})

	return agent
}

func (a *Agent) InvokeOnSessionMessage(tf *omega.TransitFrame) {
	for _, f := range a.onSessionMessageHandler {
		kkpanic.LogCatch(func() {
			f(tf)
		})
	}
}

func (a *Agent) InvokeOnRead(tf *omega.TransitFrame) {
	switch tf.GetClass() {
	case omega.TransitFrame_ClassRequest:
		for _, f := range a.onRequestHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassResponse:
		for _, f := range a.onResponseHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassRequestReplay:
		for _, f := range a.onRequestReplayHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassResponseReplay:
		for _, f := range a.onResponseReplayHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassNotification:
		if tf.GetBroadcast() != nil {
			for _, f := range a.onBroadcastHandler {
				kkpanic.LogCatch(func() {
					f(tf)
				})
			}
		}

		for _, f := range a.onNotificationHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassNotificationReplay:
		for _, f := range a.onNotificationReplayHandler {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	}

	for _, f := range a.onMessageHandler {
		kkpanic.LogCatch(func() {
			f(tf)
		})
	}
}

func (a *Agent) InvokeOnError(err error) {
	for _, f := range a.onErrorHandler {
		kkpanic.LogCatch(func() {
			f(err)
		})
	}
}

func (a *Agent) InvokeOnDisconnected() {
	for _, f := range a.onDisconnectedHandler {
		kkpanic.LogCatch(func() {
			f()
		})
	}
}

func (a *Agent) Disconnect() concurrent.Future {
	return a.session.Disconnect()
}

func (a *Agent) Session() AgentSession {
	return a.session
}

func (a *Agent) GetRemoteSession(sessionId string) RemoteSessionFuture {
	return a.session.newRemoteSession(sessionId)
}

func (a *Agent) OnDisconnected(f func()) {
	a.onDisconnectedHandler = append(a.onDisconnectedHandler, f)
}

func (a *Agent) OnMessage(f func(tf *omega.TransitFrame)) {
	a.onMessageHandler = append(a.onMessageHandler, f)
}

func (a *Agent) OnSessionMessage(f func(tf *omega.TransitFrame)) {
	a.onSessionMessageHandler = append(a.onSessionMessageHandler, f)
}

func (a *Agent) OnRequest(f func(tf *omega.TransitFrame)) {
	a.onRequestHandler = append(a.onRequestHandler, f)
}

func (a *Agent) OnResponse(f func(tf *omega.TransitFrame)) {
	a.onResponseHandler = append(a.onResponseHandler, f)
}

func (a *Agent) OnRequestReplay(f func(tf *omega.TransitFrame)) {
	a.onRequestReplayHandler = append(a.onRequestReplayHandler, f)
}

func (a *Agent) OnResponseReplay(f func(tf *omega.TransitFrame)) {
	a.onResponseReplayHandler = append(a.onResponseReplayHandler, f)
}

func (a *Agent) OnNotification(f func(tf *omega.TransitFrame)) {
	a.onNotificationHandler = append(a.onNotificationHandler, f)
}

func (a *Agent) OnNotificationReplay(f func(tf *omega.TransitFrame)) {
	a.onNotificationReplayHandler = append(a.onNotificationReplayHandler, f)
}

func (a *Agent) OnError(f func(err error)) {
	a.onErrorHandler = append(a.onErrorHandler, f)
}

func (a *Agent) Ping() concurrent.Future {
	return a.session.ping()
}

func (a *Agent) OnBroadcast(f func(tf *omega.TransitFrame)) {
	a.onBroadcastHandler = append(a.onBroadcastHandler, f)
}

func (a *Agent) Broadcast(msg string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_Broadcast{Broadcast: &omega.Broadcast{Message: msg}})
}

func (a *Agent) Hello() SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_Hello{Hello: &omega.Hello{}})
}

func (a *Agent) Time() SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_ServerTime{ServerTime: &omega.ServerTime{}})
}

func (a *Agent) PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_PlaybackChannelMessage{PlaybackChannelMessage: &omega.PlaybackChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *Agent) GetChannelMetadata(channelId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_GetChannelMeta{GetChannelMeta: &omega.GetChannelMeta{ChannelId: channelId}})
}

func (a *Agent) SetChannelMetadata(channelId string, name string, metadata *Metadata, skill *omega.Skill) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_SetChannelMeta{SetChannelMeta: &omega.SetChannelMeta{ChannelId: channelId, Name: name, Data: metadata, Skill: skill}})
}

func (a *Agent) JoinChannel(channelId string, key string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_JoinChannel{JoinChannel: &omega.JoinChannel{ChannelId: channelId, Key: key}})
}

func (a *Agent) LeaveChannel(channelId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_LeaveChannel{LeaveChannel: &omega.LeaveChannel{ChannelId: channelId}})
}

func (a *Agent) CloseChannel(channelId, key string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_CloseChannel{CloseChannel: &omega.CloseChannel{ChannelId: channelId, Key: key}})
}

func (a *Agent) ChannelMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_ChannelMessage{ChannelMessage: &omega.ChannelMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *Agent) ChannelCount(channelId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_ChannelCount{ChannelCount: &omega.ChannelCount{ChannelId: channelId}})
}

func (a *Agent) ChannelOwnerMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_ChannelOwnerMessage{ChannelOwnerMessage: &omega.ChannelOwnerMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *Agent) ReplayChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_ReplayChannelMessage{ReplayChannelMessage: &omega.ReplayChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *Agent) GetVoteMetadata(voteId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_GetVoteMeta{GetVoteMeta: &omega.GetVoteMeta{VoteId: voteId}})
}

func (a *Agent) SetVoteMetadata(voteId string, name string, metadata *Metadata) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_SetVoteMeta{SetVoteMeta: &omega.SetVoteMeta{VoteId: voteId, Name: name, Data: metadata}})
}

func (a *Agent) JoinVote(voteId string, key string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_JoinVote{JoinVote: &omega.JoinVote{VoteId: voteId, Key: key}})
}

func (a *Agent) LeaveVote(voteId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_LeaveVote{LeaveVote: &omega.LeaveVote{VoteId: voteId}})
}

func (a *Agent) CloseVote(voteId, key string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_CloseVote{CloseVote: &omega.CloseVote{VoteId: voteId, Key: key}})
}

func (a *Agent) VoteMessage(voteId string, message string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_VoteMessage{VoteMessage: &omega.VoteMessage{VoteId: voteId, Message: message}})
}

func (a *Agent) VoteSelect(voteId string, voteOptionId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_VoteSelect{VoteSelect: &omega.VoteSelect{VoteId: voteId, VoteOptionId: voteOptionId}})
}

func (a *Agent) VoteCount(voteId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_VoteCount{VoteCount: &omega.VoteCount{VoteId: voteId}})
}

func (a *Agent) VoteOwnerMessage(voteId string, message string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_VoteOwnerMessage{VoteOwnerMessage: &omega.VoteOwnerMessage{VoteId: voteId, Message: message}})
}

func (a *Agent) VoteStatus(voteId string, statusType omega.Vote_Status) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_VoteStatus{VoteStatus: &omega.VoteStatus{VoteId: voteId, Status: statusType}})
}

func (a *Agent) GetSessionMeta(sessionId string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_GetSessionMeta{GetSessionMeta: &omega.GetSessionMeta{SessionId: sessionId}})
}

func (a *Agent) SetSessionMeta(metadata *Metadata) SendFuture {
	return a.Session().SetMetadata(metadata)
}

func (a *Agent) SessionMessage(sessionId string, message string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{ToSession: sessionId, Message: message}})
}

func (a *Agent) SessionsMessage(sessionIds []string, message string) SendFuture {
	return a.session.sendRequest(&omega.TransitFrame_SessionsMessage{SessionsMessage: &omega.SessionsMessage{ToSessions: sessionIds, Message: message}})
}

func (a *Agent) CreateChannel(createChannel apirequest.CreateChannel) concurrent.Future {
	return a.Session().GetEngine().createChannel(createChannel)
}

func (a *Agent) CreateVote(createVote apirequest.CreateVote) concurrent.Future {
	return a.Session().GetEngine().createVote(createVote)
}

type AgentBuilder struct {
	config *Config
}

func NewAgentBuilder(conf *Config) *AgentBuilder {
	if conf == nil {
		return nil
	}

	return &AgentBuilder{config: conf}
}

func (b *AgentBuilder) Connect() AgentFuture {
	if b.config == nil {
		return &DefaultAgentFuture{Future: concurrent.NewFailedFuture(ErrConfigIsEmpty)}
	}

	sf := NewEngine(b.config).connect()
	af := &DefaultAgentFuture{Future: concurrent.NewFuture(), sf: sf}
	af.sf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			af.Completable().Complete(NewAgent(af.sf.Get().(*Session)))
		} else if f.IsCancelled() {
			af.Completable().Cancel()
		} else if f.IsFail() {
			af.Completable().Fail(f.Error())
		}
	}))

	return af
}
