package kumoi

import (
	"reflect"
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base"
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
	var an any
	if btf := f.bf.TransitFrame(); btf != nil {
		an = reflect.New(reflect.TypeOf(t).Elem()).Interface()
		value.Cast[messages.TransitFrameParsable](an).ParseTransitFrame(btf)
	}

	t = value.Cast[T](an)
	return
}

type RemoteSession[T base.RemoteSession] interface {
	Base() T
	GetId() string
	GetSubject() string
	GetName() string
	GetMetadata() *base.Metadata
	OnMessage(f func(msg *messages.SessionMessage))
	SendMessage(message string) SendFuture[*messages.SessionMessage]
	Fetch() SendFuture[*messages.GetSessionMeta]
}

type Session interface {
	RemoteSession[base.Session]
	SetName(name string) SendFuture[*messages.SetSessionMeta]
	SetMetadata(metadata *base.Metadata) SendFuture[*messages.SetSessionMeta]
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

func (s *remoteSession[T]) GetMetadata() *base.Metadata {
	return s.session.GetMetadata()
}

func (s *remoteSession[T]) OnMessage(f func(msg *messages.SessionMessage)) {
	s.session.OnMessage(func(msg *omega.TransitFrame) {
		if sm := msg.GetSessionMessage(); sm != nil {
			nsm := &messages.SessionMessage{}
			nsm.ParseTransitFrame(msg)
			f(nsm)
		}
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

func (s *session) SetMetadata(metadata *base.Metadata) SendFuture[*messages.SetSessionMeta] {
	return wrapSendFuture[*messages.SetSessionMeta](s.getCastSession().SetMetadata(metadata))
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
