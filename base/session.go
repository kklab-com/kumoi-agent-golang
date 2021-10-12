package base

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kklab-com/gone-core/channel"
	websocket "github.com/kklab-com/gone-websocket"
	kklogger "github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/concurrent"
	kkpanic "github.com/kklab-com/goth-panic"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	DefaultTransitTimeout        = 10 * time.Second
	DefaultKeepAlivePingInterval = time.Minute
)

var ErrConnectionClosed = errors.Errorf("connection closed")

type SendFuture interface {
	concurrent.Future
	SentTransitFrame() *omega.TransitFrame
	TransitFrame() *omega.TransitFrame
}

type DefaultSendFuture struct {
	concurrent.Future
	tf *omega.TransitFrame
}

func (f *DefaultSendFuture) SentTransitFrame() *omega.TransitFrame {
	return f.tf
}

func (f *DefaultSendFuture) TransitFrame() *omega.TransitFrame {
	if v := f.Get(); v != nil {
		return v.(*omega.TransitFrame)
	}

	return nil
}

type RemoteSession interface {
	GetId() string
	GetSubject() string
	GetName() string
	GetMetadata() *Metadata
	OnMessage(f func(msg *omega.TransitFrame))
	SendMessage(message string) SendFuture
}

type AgentSession interface {
	RemoteSession
	Ch() channel.Channel
	GetEngine() *Engine
	SetName(name string) SendFuture
	SetMetadata(metadata *Metadata) SendFuture
	Send(tf *omega.TransitFrame) SendFuture
	OnDisconnected(f func())
	OnRead(f func(tf *omega.TransitFrame))
	OnError(f func(err error))
	Disconnect() concurrent.Future
}

func newSession(engine *Engine) *Session {
	return &Session{engine: engine,
		metadata:            &structpb.Struct{},
		lastActiveTimestamp: time.Now(),
	}
}

type SessionFuture interface {
	concurrent.Future
	Session() *Session
}

type DefaultSessionFuture struct {
	concurrent.Future
}

func (f *DefaultSessionFuture) Session() *Session {
	if v := f.Get(); v != nil {
		return v.(*Session)
	}

	return nil
}

func (f *DefaultSessionFuture) Completable() concurrent.CompletableFuture {
	return f.Future.(concurrent.CompletableFuture)
}

type RemoteSessionFuture interface {
	concurrent.Future
	Session() RemoteSession
}

type DefaultRemoteSessionFuture struct {
	concurrent.Future
}

func (f *DefaultRemoteSessionFuture) Session() RemoteSession {
	if v := f.Get(); v != nil {
		return v.(*Session)
	}

	return nil
}

type Session struct {
	parent                  *Session
	serialId                int64
	engine                  *Engine
	ch                      channel.Channel
	id                      string
	subject                 string
	name                    string
	metadata                *Metadata
	connectFuture           concurrent.Future
	lastActiveTimestamp     time.Time
	onSessionMessageHandler []func(msg *omega.TransitFrame)
	onReadHandler           []func(tf *omega.TransitFrame)
	onDisconnectedHandler   []func()
	onErrorHandler          []func(err error)
	transitPool             sync.Map
}

func (s *Session) newRemoteSession(sessionId string) RemoteSessionFuture {
	rsf := &DefaultRemoteSessionFuture{concurrent.NewFuture(nil)}
	f := s.sendRequest(&omega.TransitFrame_GetSessionMeta{GetSessionMeta: &omega.GetSessionMeta{SessionId: sessionId}})
	f.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			if v := f.Get(); v == nil {
				rsf.Completable().Fail(ErrUnexpectError)
				return
			}

			if sm := f.Get().(*omega.TransitFrame).GetGetSessionMeta(); sm != nil {
				rs := &Session{
					parent:   s,
					ch:       s.ch,
					id:       sm.GetSessionId(),
					subject:  sm.GetSubject(),
					name:     sm.GetName(),
					metadata: sm.GetData(),
				}

				s.OnMessage(func(tf *omega.TransitFrame) {
					if sm := tf.GetSessionMessage(); sm != nil && sm.GetFromSession() == rs.id {
						for _, f := range rs.onSessionMessageHandler {
							kkpanic.LogCatch(func() {
								f(tf)
							})
						}
					}
				})

				rsf.Completable().Complete(rs)
			} else {
				rsf.Completable().Fail(ErrUnexpectError)
			}
		} else if f.IsCancelled() {
			rsf.Completable().Cancel()
		} else if f.IsError() {
			rsf.Completable().Fail(ErrSessionNotFound)
		}
	}))

	return rsf
}

func (s *Session) Ch() channel.Channel {
	return s.ch
}

func (s *Session) GetEngine() *Engine {
	return s.engine
}

func (s *Session) GetId() string {
	return s.id
}

func (s *Session) GetSubject() string {
	return s.subject
}

func (s *Session) GetName() string {
	return s.name
}

func (s *Session) SetName(name string) SendFuture {
	uname := name
	rf := s.sendRequest(
		&omega.TransitFrame_SetSessionMeta{SetSessionMeta: &omega.SetSessionMeta{
			Data: s.GetMetadata(),
			Name: uname,
		}})

	rf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			if tf, ok := f.Get().(*omega.TransitFrame); ok && tf.GetClass() == omega.TransitFrame_ClassResponse {
				s.name = uname
			}
		}
	}))

	return rf
}

func (s *Session) GetMetadata() *Metadata {
	return s.metadata
}

func (s *Session) SetMetadata(metadata *Metadata) SendFuture {
	uMetadata := metadata
	rf := s.sendRequest(
		&omega.TransitFrame_SetSessionMeta{SetSessionMeta: &omega.SetSessionMeta{
			Data: uMetadata,
			Name: s.GetName(),
		}})

	rf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsSuccess() {
			if tf, ok := f.Get().(*omega.TransitFrame); ok && tf.GetClass() == omega.TransitFrame_ClassResponse {
				s.metadata = uMetadata
			}
		}
	}))

	return rf
}

func (s *Session) ping() SendFuture {
	return s.sendRequest(&omega.TransitFrame_Ping{Ping: &omega.Ping{}})
}

// OnMessage
// for SessionMessage
func (s *Session) OnMessage(f func(msg *omega.TransitFrame)) {
	s.onSessionMessageHandler = append(s.onSessionMessageHandler, f)
}

func (s *Session) SendMessage(message string) SendFuture {
	return s.sendRequest(
		&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{
			ToSession: s.GetId(),
			Message:   message,
		}})
}

func (s *Session) NewTransitId() uint64 {
	if s.parent == nil {
		return uint64(atomic.AddInt64(&s.serialId, 1))
	} else {
		return uint64(atomic.AddInt64(&s.parent.serialId, 1))
	}
}

func (s *Session) newTransitFrame(class omega.TransitFrame_FrameClass, data omega.TransitFrameData) *omega.TransitFrame {
	return &omega.TransitFrame{
		TransitId: s.NewTransitId(),
		Timestamp: time.Now().UnixNano(),
		Class:     class,
		Version:   omega.TransitFrame_VersionBase,
		Data:      data,
	}
}

func (s *Session) OnDisconnected(f func()) {
	s.onDisconnectedHandler = append(s.onDisconnectedHandler, f)
}

// OnRead
// for all message
func (s *Session) OnRead(f func(tf *omega.TransitFrame)) {
	s.onReadHandler = append(s.onReadHandler, f)
}

func (s *Session) OnError(f func(err error)) {
	s.onErrorHandler = append(s.onErrorHandler, f)
}

func (s *Session) invokeOnRead(tf *omega.TransitFrame) {
	if tf == nil {
		kklogger.ErrorJ("Session.invokeOnRead", "nil tf")
		return
	}

	switch tf.GetClass() {
	case omega.TransitFrame_ClassRequest:
	case omega.TransitFrame_ClassResponse:
		if v, f := s.getSourceSession().transitPool.LoadAndDelete(tf.GetTransitId()); f {
			v.(*transitPoolEntity).future.Completable().Complete(tf)
		}
	case omega.TransitFrame_ClassNotification:
		if tf.GetSessionMessage() != nil {
			for _, f := range s.onSessionMessageHandler {
				kkpanic.LogCatch(func() {
					f(tf)
				})
			}
		}
	case omega.TransitFrame_ClassError:
		if v, f := s.getSourceSession().transitPool.LoadAndDelete(tf.GetTransitId()); f {
			v.(*transitPoolEntity).future.Completable().Fail(tf)
		}
	}

	for _, f := range s.onReadHandler {
		kkpanic.LogCatch(func() {
			f(tf)
		})
	}
}

func (s *Session) Send(tf *omega.TransitFrame) SendFuture {
	stf := tf
	if s.isClosed() {
		rf := &DefaultSendFuture{
			Future: concurrent.NewFailedFuture(ErrConnectionClosed),
			tf:     stf,
		}

		return rf
	}

	bs, _ := proto.Marshal(stf)
	sf := s.ch.Write(&websocket.DefaultMessage{
		MessageType: websocket.BinaryMessageType,
		Message:     bs,
	})

	var rf SendFuture
	rf = &DefaultSendFuture{
		Future: concurrent.NewFuture(nil),
		tf:     stf,
	}

	sf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsError() {
			rf.Completable().Fail(f.Error())
		}
	}))

	if stf.GetClass() == omega.TransitFrame_ClassRequest {
		s.getSourceSession().transitPool.Store(stf.GetTransitId(), &transitPoolEntity{
			timestamp: time.Now(),
			future:    rf,
		})
	} else {
		sf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
			if f.IsSuccess() {
				rf.Completable().Complete(nil)
			}
		}))
	}

	s.lastActiveTimestamp = time.Now()
	return rf
}

func (s *Session) getSourceSession() *Session {
	if p := s.parent; p != nil {
		return p
	}

	return s
}

func (s *Session) sendRequest(data omega.TransitFrameData) SendFuture {
	return s.Send(s.newTransitFrame(omega.TransitFrame_ClassRequest, data))
}

func (s *Session) Disconnect() concurrent.Future {
	return s.ch.Disconnect()
}

func (s *Session) isClosed() bool {
	return !s.ch.IsActive()
}
