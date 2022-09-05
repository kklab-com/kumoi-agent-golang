package kumoi

import (
	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type RemoteSession interface {
	Base() base.RemoteSession
	GetId() string
	GetSubject() string
	GetName() string
	GetMetadata() *base.Metadata
	OnMessage(f func(msg *messages.SessionMessage))
	SendMessage(message string) base.SendFuture
}

type AgentSession interface {
	Base() base.AgentSession
	GetId() string
	GetSubject() string
	GetName() string
	GetMetadata() *base.Metadata
	OnMessage(f func(msg *messages.SessionMessage))
	SendMessage(message string) base.SendFuture
	SetName(name string) base.SendFuture
	SetMetadata(metadata *base.Metadata) base.SendFuture
	OnDisconnected(f func())
	OnError(f func(err error))
	Disconnect() concurrent.Future
}

type remoteSession struct {
	session base.RemoteSession
}

func (r *remoteSession) Base() base.RemoteSession {
	return r.session
}

func (r *remoteSession) GetId() string {
	return r.session.GetId()
}

func (r *remoteSession) GetSubject() string {
	return r.session.GetSubject()
}

func (r *remoteSession) GetName() string {
	return r.session.GetName()
}

func (r *remoteSession) GetMetadata() *base.Metadata {
	return r.session.GetMetadata()
}

func (r *remoteSession) OnMessage(f func(msg *messages.SessionMessage)) {
	r.session.OnMessage(func(msg *omega.TransitFrame) {
		if sm := msg.GetSessionMessage(); sm != nil {
			nsm := &messages.SessionMessage{}
			nsm.ParseTransitFrame(msg)
			f(nsm)
		}
	})
}

func (r *remoteSession) SendMessage(message string) base.SendFuture {
	return r.session.SendMessage(message)
}

type agentSession struct {
	session base.AgentSession
}

func (r *agentSession) Base() base.AgentSession {
	return r.session
}

func (r *agentSession) GetId() string {
	return r.session.GetId()
}

func (r *agentSession) GetSubject() string {
	return r.session.GetSubject()
}

func (r *agentSession) GetName() string {
	return r.session.GetName()
}

func (r *agentSession) GetMetadata() *base.Metadata {
	return r.session.GetMetadata()
}

func (r *agentSession) OnMessage(f func(msg *messages.SessionMessage)) {
	r.session.OnMessage(func(msg *omega.TransitFrame) {
		if sm := msg.GetSessionMessage(); sm != nil {
			nsm := &messages.SessionMessage{}
			nsm.ParseTransitFrame(msg)
			f(nsm)
		}
	})
}

func (r *agentSession) SendMessage(message string) base.SendFuture {
	return r.session.SendMessage(message)
}

func (r *agentSession) SetName(name string) base.SendFuture {
	return r.session.SetName(name)
}

func (r *agentSession) SetMetadata(metadata *base.Metadata) base.SendFuture {
	return r.session.SetMetadata(metadata)
}

func (r *agentSession) OnDisconnected(f func()) {
	r.session.OnDisconnected(f)
}

func (r *agentSession) OnError(f func(err error)) {
	r.session.OnError(f)
}

func (r *agentSession) Disconnect() concurrent.Future {
	return r.session.Disconnect()
}
