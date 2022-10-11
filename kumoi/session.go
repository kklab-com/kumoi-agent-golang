package kumoi

import (
	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type RemoteSession[T base.RemoteSession] interface {
	Base() T
	GetId() string
	GetSubject() string
	GetName() string
	GetMetadata() map[string]any
	OnMessage(f func(msg *messages.SessionMessage))
	SendMessage(message string) SendFuture[*messages.SessionMessage]
	Fetch() SendFuture[*messages.GetSessionMeta]
}

type Session interface {
	RemoteSession[base.Session]
	SetName(name string) SendFuture[*messages.SetSessionMeta]
	SetMetadata(metadata map[string]any) SendFuture[*messages.SetSessionMeta]
	OnRead(f func(th *omega.TransitFrame))
	OnClosed(f func())
	OnError(f func(err error))
	Close() concurrent.Future
}

type remoteSession[T base.RemoteSession] struct {
	session base.RemoteSession
}

func (s *remoteSession[T]) Base() T {
	return value.Cast[T](s.session)
}

func (s *remoteSession[T]) GetId() string {
	return s.session.GetId()
}

func (s *remoteSession[T]) GetSubject() string {
	return s.session.GetSubject()
}

func (s *remoteSession[T]) GetName() string {
	return s.session.GetName()
}

func (s *remoteSession[T]) GetMetadata() map[string]any {
	return base.SafeGetStructMap(s.session.GetMetadata())
}

func (s *remoteSession[T]) OnMessage(f func(msg *messages.SessionMessage)) {
	s.session.OnMessage(func(msg *omega.TransitFrame) {
		f(value.Cast[*messages.SessionMessage](messages.WrapTransitFrame(msg)))
	})
}

func (s *remoteSession[T]) SendMessage(message string) SendFuture[*messages.SessionMessage] {
	return wrapSendFuture[*messages.SessionMessage](s.session.SendMessage(message))
}

func (s *remoteSession[T]) Fetch() SendFuture[*messages.GetSessionMeta] {
	return wrapSendFuture[*messages.GetSessionMeta](s.session.Fetch())
}

type session struct {
	remoteSession[base.Session]
}

func (s *session) getCastSession() base.Session {
	return value.Cast[base.Session](s.remoteSession.session)
}

func (s *session) SetName(name string) SendFuture[*messages.SetSessionMeta] {
	return wrapSendFuture[*messages.SetSessionMeta](s.getCastSession().SetName(name))
}

func (s *session) SetMetadata(metadata map[string]any) SendFuture[*messages.SetSessionMeta] {
	return wrapSendFuture[*messages.SetSessionMeta](s.getCastSession().SetMetadata(base.NewMetadata(metadata)))
}

func (s *session) OnRead(f func(th *omega.TransitFrame)) {
	s.getCastSession().OnRead(f)
}

func (s *session) OnClosed(f func()) {
	s.getCastSession().OnClosed(f)
}

func (s *session) OnError(f func(err error)) {
	s.getCastSession().OnError(f)
}

func (s *session) Close() concurrent.Future {
	return s.session.Close()
}
