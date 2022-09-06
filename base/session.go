package base

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/kklab-com/gone-core/channel"
	websocket "github.com/kklab-com/gone-websocket"
	"github.com/kklab-com/goth-concurrent"
	kklogger "github.com/kklab-com/goth-kklogger"
	kkpanic "github.com/kklab-com/goth-panic"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
	"github.com/pkg/errors"
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
	Close() concurrent.Future
}

type Session interface {
	RemoteSession
	GetRemoteSession(sessionId string) RemoteSessionFuture
	Ch() channel.Channel
	GetEngine() *Engine
	SetName(name string) SendFuture
	SetMetadata(metadata *Metadata) SendFuture
	Ping() SendFuture
	Send(tf *omega.TransitFrame) SendFuture
	SendRequest(data omega.TransitFrameData) SendFuture
	OnClosed(f func())
	OnRead(f func(tf *omega.TransitFrame))
	OnError(f func(err error))
	IsClosed() bool
}

type SessionFuture interface {
	concurrent.Future
	Session() Session
}

type DefaultSessionFuture struct {
	concurrent.Future
}

func (f *DefaultSessionFuture) Session() Session {
	if v := f.Get(); v != nil {
		return v.(*session)
	}

	return nil
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
		return v.(*remoteSession)
	}

	return nil
}

type remoteSession struct {
	source                  *session
	id                      string
	subject                 string
	name                    string
	metadata                *Metadata
	onSessionMessageHandler func(msg *omega.TransitFrame)
}

func (s *remoteSession) GetId() string {
	return s.id
}

func (s *remoteSession) GetSubject() string {
	return s.subject
}

func (s *remoteSession) GetName() string {
	return s.name
}

func (s *remoteSession) GetMetadata() *Metadata {
	return s.metadata
}

// OnMessage
// for SessionMessage
func (s *remoteSession) OnMessage(f func(msg *omega.TransitFrame)) {
	s.onSessionMessageHandler = f
}

func (s *remoteSession) SendMessage(message string) SendFuture {
	return s.source.SendRequest(
		&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{
			ToSession: s.GetId(),
			Message:   message,
		}})
}

func (s *remoteSession) Close() concurrent.Future {
	if s.source == nil || s.source.IsClosed() {
		return concurrent.NewFailedFuture(ErrConnectionClosed)
	}

	s.source.remoteSessions.Delete(s.GetId())
	return concurrent.NewCompletedFuture(nil)
}

func newSession(engine *Engine) *session {
	return &session{engine: engine,
		remoteSession: remoteSession{
			metadata: NewMetadata(nil),
		},
		transitFrameWorkQueue: concurrent.NewUnlimitedBlockingQueue(),
		lastActiveTimestamp:   time.Now(),
	}
}

type session struct {
	remoteSession
	serialId                 int64
	engine                   *Engine
	ch                       channel.Channel
	connectFuture            concurrent.Future
	lastActiveTimestamp      time.Time
	workerStatus             int32
	transitFrameWorkQueue    concurrent.BlockingQueue
	onSessionMessageHandlers []func(msg *omega.TransitFrame)
	onReadHandlers           []func(tf *omega.TransitFrame)
	onClosedHandlers         []func()
	onErrorHandlers          []func(err error)
	transitPool              sync.Map
	remoteSessions           sync.Map
}

func (s *session) GetRemoteSession(sessionId string) RemoteSessionFuture {
	rsf := &DefaultRemoteSessionFuture{concurrent.NewFuture()}
	source := s
	source.SendRequest(&omega.TransitFrame_GetSessionMeta{GetSessionMeta: &omega.GetSessionMeta{SessionId: sessionId}}).
		AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
			if f.IsSuccess() {
				if v := f.Get(); v == nil {
					rsf.Completable().Fail(ErrUnexpectError)
					return
				}

				if sm := f.Get().(*omega.TransitFrame).GetGetSessionMeta(); sm != nil {
					rs := &remoteSession{
						source:   source,
						id:       sm.GetSessionId(),
						subject:  sm.GetSubject(),
						name:     sm.GetName(),
						metadata: sm.GetData(),
					}

					source.remoteSessions.Store(rs.GetId(), rs)
					rsf.Completable().Complete(rs)
				} else {
					rsf.Completable().Fail(ErrUnexpectError)
				}
			} else if f.IsCancelled() {
				rsf.Completable().Cancel()
			} else if f.IsFail() {
				rsf.Completable().Fail(ErrSessionNotFound)
			}
		}))

	return rsf
}

func (s *session) Ch() channel.Channel {
	return s.ch
}

func (s *session) GetEngine() *Engine {
	return s.engine
}

func (s *session) SetName(name string) SendFuture {
	uname := name
	rf := s.SendRequest(
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

func (s *session) SetMetadata(metadata *Metadata) SendFuture {
	uMetadata := metadata
	rf := s.SendRequest(
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

func (s *session) Ping() SendFuture {
	return s.SendRequest(&omega.TransitFrame_Ping{Ping: &omega.Ping{}})
}

func (s *session) Send(tf *omega.TransitFrame) SendFuture {
	stf := tf
	if s.IsClosed() {
		return &DefaultSendFuture{
			Future: concurrent.NewFailedFuture(ErrConnectionClosed),
			tf:     stf,
		}
	}

	var rsf SendFuture
	rsf = &DefaultSendFuture{
		Future: concurrent.NewFuture(),
		tf:     stf,
	}

	if stf.GetClass() == omega.TransitFrame_ClassRequest {
		s.transitPool.Store(stf.GetTransitId(), &transitPoolEntity{
			timestamp: time.Now(),
			future:    rsf,
		})
	}

	bs, _ := proto.Marshal(stf)
	sf := s.ch.Write(&websocket.DefaultMessage{
		MessageType: websocket.BinaryMessageType,
		Message:     bs,
	})

	sf.AddListener(concurrent.NewFutureListener(func(f concurrent.Future) {
		if f.IsFail() {
			rsf.Completable().Fail(f.Error())
		} else if f.IsCancelled() {
			rsf.Completable().Cancel()
		}

		if stf.GetClass() != omega.TransitFrame_ClassRequest && f.IsSuccess() {
			rsf.Completable().Complete(nil)
		}
	}))

	s.lastActiveTimestamp = time.Now()
	return rsf
}

func (s *session) SendRequest(data omega.TransitFrameData) SendFuture {
	return s.Send(s.NewTransitFrame(omega.TransitFrame_ClassRequest, data))
}

func (s *session) OnClosed(f func()) {
	s.onClosedHandlers = append(s.onClosedHandlers, f)
}

// OnRead
// for all message
func (s *session) OnRead(f func(tf *omega.TransitFrame)) {
	s.onReadHandlers = append(s.onReadHandlers, f)
}

func (s *session) OnError(f func(err error)) {
	s.onErrorHandlers = append(s.onErrorHandlers, f)
}

// OnMessage
// for SessionMessage
func (s *session) OnMessage(f func(msg *omega.TransitFrame)) {
	s.onSessionMessageHandlers = append(s.onSessionMessageHandlers, f)
}

func (s *session) SendMessage(message string) SendFuture {
	return s.SendRequest(
		&omega.TransitFrame_SessionMessage{SessionMessage: &omega.SessionMessage{
			ToSession: s.GetId(),
			Message:   message,
		}})
}

func (s *session) newTransitId() uint64 {
	return uint64(atomic.AddInt64(&s.serialId, 1))
}

func (s *session) NewTransitFrame(class omega.TransitFrame_FrameClass, data omega.TransitFrameData) *omega.TransitFrame {
	return &omega.TransitFrame{
		TransitId: s.newTransitId(),
		Timestamp: time.Now().UnixNano(),
		Class:     class,
		Version:   omega.TransitFrame_VersionBase,
		Data:      data,
	}
}

func (s *session) transitFramePreProcess(tf *omega.TransitFrame) {
	// hello msg
	if hello := tf.GetHello(); hello != nil {
		s.subject = hello.GetSubject()
		s.name = hello.GetSubjectName()
		s.id = hello.GetSessionId()
		s.connectFuture.Completable().Complete(s)
	}

	// ping, auto pong reply
	if ping := tf.GetPing(); ping != nil && tf.GetClass() == omega.TransitFrame_ClassRequest {
		s.Send(tf.Clone().AsResponse().RenewTimestamp().SetData(&omega.TransitFrame_Pong{Pong: &omega.Pong{}}))
	}
}

func (s *session) completeWhenResponseAndError(tf *omega.TransitFrame) {
	if tf.GetClass() == omega.TransitFrame_ClassResponse {
		if v, f := s.transitPool.LoadAndDelete(tf.GetTransitId()); f {
			v.(*transitPoolEntity).future.Completable().Complete(tf)
		}
	} else if tf.GetClass() == omega.TransitFrame_ClassError {
		if v, f := s.transitPool.LoadAndDelete(tf.GetTransitId()); f {
			v.(*transitPoolEntity).future.Completable().Fail(tf)
		}
	}
}

func (s *session) invokeOnRead(tf *omega.TransitFrame) {
	if tf == nil {
		kklogger.ErrorJ("session.invokeOnRead", "nil tf")
		return
	}

	s.completeWhenResponseAndError(tf)
	s.submitTransitFrameWorker(tf)
}

func (s *session) submitTransitFrameWorker(tf *omega.TransitFrame) {
	s.transitFrameWorkQueue.Push(tf)
	if atomic.CompareAndSwapInt32(&s.workerStatus, 0, 1) {
		go s.transitFrameWorker(true)
	}
}

func (s *session) transitFrameWorker(retry bool) {
	for !s.IsClosed() {
		if v := s.transitFrameWorkQueue.TryPop(); v != nil {
			if !retry {
				retry = !retry
			}

			if tf, ok := v.(*omega.TransitFrame); ok {
				for _, f := range s.onReadHandlers {
					kkpanic.LogCatch(func() {
						f(tf)
					})
				}

				if sm := tf.GetSessionMessage(); sm != nil {
					if tf.GetClass() == omega.TransitFrame_ClassNotification {
						for _, f := range s.onSessionMessageHandlers {
							kkpanic.LogCatch(func() {
								f(tf)
							})
						}
					}

					if v, f := s.remoteSessions.Load(sm.GetFromSession()); f {
						v.(*remoteSession).onSessionMessageHandler(tf)
					}
				}
			}
		} else {
			if retry {
				atomic.CompareAndSwapInt32(&s.workerStatus, 1, 0)
				s.transitFrameWorker(false)
			}

			break
		}
	}
}

func (s *session) Close() concurrent.Future {
	return s.ch.Disconnect()
}

func (s *session) IsClosed() bool {
	return !s.ch.IsActive()
}
