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
}

func (f *DefaultAgentFuture) Agent() *Agent {
	if v := f.Get(); v != nil {
		return v.(*Agent)
	}

	return nil
}

type Agent struct {
	session                      Session
	onClosedHandlers             []func()
	onMessageHandlers            []func(tf *omega.TransitFrame)
	onSessionMessageHandlers     []func(tf *omega.TransitFrame)
	onBroadcastHandlers          []func(tf *omega.TransitFrame)
	onRequestHandlers            []func(tf *omega.TransitFrame)
	onResponseHandlers           []func(tf *omega.TransitFrame)
	onRequestReplayHandlers      []func(tf *omega.TransitFrame)
	onResponseReplayHandlers     []func(tf *omega.TransitFrame)
	onNotificationHandlers       []func(tf *omega.TransitFrame)
	onNotificationReplayHandlers []func(tf *omega.TransitFrame)
	onErrorHandlers              []func(err error)
}

func NewAgent(session Session) *Agent {
	if session == nil {
		return nil
	}

	agent := &Agent{
		session: session,
		onErrorHandlers: []func(err error){func(err error) {
			kklogger.ErrorJ("base:Agent.onErrorHandlers", err.Error())
		}},
	}

	session.OnMessage(func(msg *omega.TransitFrame) {
		agent.invokeOnSessionMessage(msg)
	})

	session.OnRead(func(tf *omega.TransitFrame) {
		agent.invokeOnRead(tf)
	})

	session.OnClosed(func() {
		agent.invokeOnClosed()
	})

	session.OnError(func(err error) {
		agent.invokeOnError(err)
	})

	return agent
}

func (a *Agent) invokeOnSessionMessage(tf *omega.TransitFrame) {
	for _, f := range a.onSessionMessageHandlers {
		kkpanic.LogCatch(func() {
			f(tf)
		})
	}
}

func (a *Agent) invokeOnRead(tf *omega.TransitFrame) {
	switch tf.GetClass() {
	case omega.TransitFrame_ClassRequest:
		for _, f := range a.onRequestHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassResponse:
		for _, f := range a.onResponseHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassRequestReplay:
		for _, f := range a.onRequestReplayHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassResponseReplay:
		for _, f := range a.onResponseReplayHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassNotification:
		if tf.GetBroadcast() != nil {
			for _, f := range a.onBroadcastHandlers {
				kkpanic.LogCatch(func() {
					f(tf)
				})
			}
		}

		for _, f := range a.onNotificationHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	case omega.TransitFrame_ClassNotificationReplay:
		for _, f := range a.onNotificationReplayHandlers {
			kkpanic.LogCatch(func() {
				f(tf)
			})
		}
	}

	for _, f := range a.onMessageHandlers {
		kkpanic.LogCatch(func() {
			f(tf)
		})
	}
}

func (a *Agent) invokeOnError(err error) {
	for _, f := range a.onErrorHandlers {
		kkpanic.LogCatch(func() {
			f(err)
		})
	}
}

func (a *Agent) invokeOnClosed() {
	for _, f := range a.onClosedHandlers {
		kkpanic.LogCatch(func() {
			f()
		})
	}
}

func (a *Agent) Close() concurrent.Future {
	return a.session.Close()
}

func (a *Agent) Session() Session {
	return a.session
}

func (a *Agent) GetRemoteSession(sessionId string) RemoteSessionFuture {
	return a.session.GetRemoteSession(sessionId)
}

func (a *Agent) OnClosed(f func()) {
	a.onClosedHandlers = append(a.onClosedHandlers, f)
}

func (a *Agent) OnMessage(f func(tf *omega.TransitFrame)) {
	a.onMessageHandlers = append(a.onMessageHandlers, f)
}

func (a *Agent) OnSessionMessage(f func(tf *omega.TransitFrame)) {
	a.onSessionMessageHandlers = append(a.onSessionMessageHandlers, f)
}

func (a *Agent) OnRequest(f func(tf *omega.TransitFrame)) {
	a.onRequestHandlers = append(a.onRequestHandlers, f)
}

func (a *Agent) OnResponse(f func(tf *omega.TransitFrame)) {
	a.onResponseHandlers = append(a.onResponseHandlers, f)
}

func (a *Agent) OnRequestReplay(f func(tf *omega.TransitFrame)) {
	a.onRequestReplayHandlers = append(a.onRequestReplayHandlers, f)
}

func (a *Agent) OnResponseReplay(f func(tf *omega.TransitFrame)) {
	a.onResponseReplayHandlers = append(a.onResponseReplayHandlers, f)
}

func (a *Agent) OnNotification(f func(tf *omega.TransitFrame)) {
	a.onNotificationHandlers = append(a.onNotificationHandlers, f)
}

func (a *Agent) OnNotificationReplay(f func(tf *omega.TransitFrame)) {
	a.onNotificationReplayHandlers = append(a.onNotificationReplayHandlers, f)
}

func (a *Agent) OnError(f func(err error)) {
	a.onErrorHandlers = append(a.onErrorHandlers, f)
}

func (a *Agent) Ping() concurrent.Future {
	return a.session.Ping()
}

func (a *Agent) OnBroadcast(f func(tf *omega.TransitFrame)) {
	a.onBroadcastHandlers = append(a.onBroadcastHandlers, f)
}

func (a *Agent) Broadcast(msg string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_Broadcast{Broadcast: &omega.Broadcast{Message: msg}})
}

func (a *Agent) Hello() SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_Hello{Hello: &omega.Hello{}})
}

func (a *Agent) Time() SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ServerTime{ServerTime: &omega.ServerTime{}})
}

func (a *Agent) PlaybackChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_PlaybackChannelMessage{PlaybackChannelMessage: &omega.PlaybackChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *Agent) GetChannelMetadata(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetChannelMeta{GetChannelMeta: &omega.GetChannelMeta{ChannelId: channelId}})
}

func (a *Agent) SetChannelMetadata(channelId string, name string, metadata *Metadata, skill *omega.Skill) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SetChannelMeta{SetChannelMeta: &omega.SetChannelMeta{ChannelId: channelId, Name: name, Data: metadata, Skill: skill}})
}

func (a *Agent) JoinChannel(channelId string, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_JoinChannel{JoinChannel: &omega.JoinChannel{ChannelId: channelId, Key: key}})
}

func (a *Agent) LeaveChannel(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_LeaveChannel{LeaveChannel: &omega.LeaveChannel{ChannelId: channelId}})
}

func (a *Agent) CloseChannel(channelId, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_CloseChannel{CloseChannel: &omega.CloseChannel{ChannelId: channelId, Key: key}})
}

func (a *Agent) ChannelMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelMessage{ChannelMessage: &omega.ChannelMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *Agent) ChannelCount(channelId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelCount{ChannelCount: &omega.ChannelCount{ChannelId: channelId}})
}

func (a *Agent) ChannelOwnerMessage(channelId string, message string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ChannelOwnerMessage{ChannelOwnerMessage: &omega.ChannelOwnerMessage{ChannelId: channelId, Message: message, Metadata: metadata}})
}

func (a *Agent) ReplayChannelMessage(channelId string, targetTimestamp int64, inverse bool, volume omega.Volume, nextId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_ReplayChannelMessage{ReplayChannelMessage: &omega.ReplayChannelMessage{
		ChannelId: channelId,
		EndedAt:   targetTimestamp,
		NextId:    nextId,
		Volume:    volume,
		Inverse:   inverse,
	}})
}

func (a *Agent) GetVoteMetadata(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetVoteMeta{GetVoteMeta: &omega.GetVoteMeta{VoteId: voteId}})
}

func (a *Agent) SetVoteMetadata(voteId string, name string, metadata *Metadata) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SetVoteMeta{SetVoteMeta: &omega.SetVoteMeta{VoteId: voteId, Name: name, Data: metadata}})
}

func (a *Agent) JoinVote(voteId string, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_JoinVote{JoinVote: &omega.JoinVote{VoteId: voteId, Key: key}})
}

func (a *Agent) LeaveVote(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_LeaveVote{LeaveVote: &omega.LeaveVote{VoteId: voteId}})
}

func (a *Agent) CloseVote(voteId, key string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_CloseVote{CloseVote: &omega.CloseVote{VoteId: voteId, Key: key}})
}

func (a *Agent) VoteMessage(voteId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteMessage{VoteMessage: &omega.VoteMessage{VoteId: voteId, Message: message}})
}

func (a *Agent) VoteSelect(voteId string, voteOptionId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteSelect{VoteSelect: &omega.VoteSelect{VoteId: voteId, VoteOptionId: voteOptionId}})
}

func (a *Agent) VoteCount(voteId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteCount{VoteCount: &omega.VoteCount{VoteId: voteId}})
}

func (a *Agent) VoteOwnerMessage(voteId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteOwnerMessage{VoteOwnerMessage: &omega.VoteOwnerMessage{VoteId: voteId, Message: message}})
}

func (a *Agent) VoteStatus(voteId string, statusType omega.Vote_Status) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_VoteStatus{VoteStatus: &omega.VoteStatus{VoteId: voteId, Status: statusType}})
}

func (a *Agent) GetSessionMeta(sessionId string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_GetSessionMeta{GetSessionMeta: &omega.GetSessionMeta{SessionId: sessionId}})
}

func (a *Agent) SetSessionMeta(metadata *Metadata) SendFuture {
	return a.Session().SetMetadata(metadata)
}

func (a *Agent) SessionMessage(sessionId string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{ToSession: sessionId, Message: message}})
}

func (a *Agent) SessionsMessage(sessionIds []string, message string) SendFuture {
	return a.session.SendRequest(&omega.TransitFrame_SessionsMessage{SessionsMessage: &omega.SessionsMessage{ToSessions: sessionIds, Message: message}})
}

func (a *Agent) CreateChannel(createChannel apirequest.CreateChannel) concurrent.Future {
	return a.Session().GetEngine().createChannel(createChannel)
}

func (a *Agent) CreateVote(createVote apirequest.CreateVote) concurrent.Future {
	return a.Session().GetEngine().createVote(createVote)
}

type AgentBuilder struct {
	engine *Engine
}

func NewAgentBuilder(engine *Engine) *AgentBuilder {
	if engine == nil {
		return nil
	}

	return &AgentBuilder{engine: engine}
}

func (b *AgentBuilder) Connect() AgentFuture {
	if b.engine == nil || b.engine.Config == nil {
		return &DefaultAgentFuture{Future: concurrent.NewFailedFuture(ErrConfigIsEmpty)}
	}

	af := &DefaultAgentFuture{Future: concurrent.NewFuture()}
	b.engine.connect().AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			af.Completable().Complete(NewAgent(f.Get().(*session)))
		} else if f.IsCancelled() {
			af.Completable().Cancel()
		} else {
			af.Completable().Fail(f.Error())
		}
	}))

	return af
}
