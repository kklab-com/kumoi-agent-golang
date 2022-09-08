package messages

import (
	"reflect"

	"github.com/kklab-com/goth-kkutil/value"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

var TransitMessageIdCodec = &omega.TransitMessageIDCodec{}

type TransitFrame interface {
	GetTransitId() uint64
	GetClass() omega.TransitFrame_FrameClass
	GetVersion() omega.TransitFrame_Version
	GetTimestamp() int64
	GetErr() *omega.Error
	GetMessageId() string
	GetRefererMessageId() string
	BaseTransitFrame() *omega.TransitFrame
	ProtoMessage() *omega.TransitFrame
	TypeName() string
	Cast() CastTransitFrame
}

type TransitFrameParsable interface {
	ParseTransitFrame(tf *omega.TransitFrame)
}

type transitFrame struct {
	tf   *omega.TransitFrame
	cast castTransitFrame
}

func (m *transitFrame) setCast(tf TransitFrame) {
	m.cast.tf = tf
}

func (m *transitFrame) GetTransitId() uint64 {
	return m.tf.GetTransitId()
}

func (m *transitFrame) GetClass() omega.TransitFrame_FrameClass {
	return m.tf.GetClass()
}

func (m *transitFrame) GetVersion() omega.TransitFrame_Version {
	return m.tf.GetVersion()
}

func (m *transitFrame) GetTimestamp() int64 {
	return m.tf.GetTimestamp()
}

func (m *transitFrame) GetErr() *omega.Error {
	return m.tf.GetErr()
}

func (m *transitFrame) GetMessageId() string {
	return TransitMessageIdCodec.EncodeToString(m.tf.GetMessageId())
}

func (m *transitFrame) GetRefererMessageId() string {
	return TransitMessageIdCodec.EncodeToString(m.tf.GetRefererMessageId())
}

func (m *transitFrame) BaseTransitFrame() *omega.TransitFrame {
	return m.tf
}

func (m *transitFrame) ProtoMessage() *omega.TransitFrame {
	return m.BaseTransitFrame()
}

func (m *transitFrame) TypeName() string {
	tfd := reflect.ValueOf(m.ProtoMessage().GetData())
	if !tfd.IsValid() {
		return ""
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return ""
	}

	return tfdE.Field(0).Elem().Type().Name()
}

func (m *transitFrame) Cast() CastTransitFrame {
	return &m.cast
}

func (m *transitFrame) Error() string {
	if err := m.GetErr(); err != nil {
		return value.JsonMarshal(m.GetErr())
	}

	return ""
}

func (m *transitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	m.tf = tf
}

type CastTransitFrame interface {
	Hello() *Hello
	ServerTime() *ServerTime
	ChannelFrame() ChannelFrame
	JoinChannel() *JoinChannel
	GetChannelMeta() *GetChannelMeta
	SetChannelMeta() *SetChannelMeta
	ChannelMessage() *ChannelMessage
	ChannelOwnerMessage() *ChannelOwnerMessage
	ChannelCount() *ChannelCount
	LeaveChannel() *LeaveChannel
	CloseChannel() *CloseChannel
	VoteFrame() VoteFrame
	JoinVote() *JoinVote
	GetVoteMeta() *GetVoteMeta
	SetVoteMeta() *SetVoteMeta
	VoteMessage() *VoteMessage
	VoteOwnerMessage() *VoteOwnerMessage
	VoteCount() *VoteCount
	VoteSelect() *VoteSelect
	VoteStatus() *VoteStatus
	LeaveVote() *LeaveVote
	CloseVote() *CloseVote
	GetSessionMeta() *GetSessionMeta
	SetSessionMeta() *SetSessionMeta
	SessionMessage() *SessionMessage
}

type castTransitFrame struct {
	tf TransitFrame
}

func (m *castTransitFrame) Hello() *Hello {
	return value.Cast[*Hello](m.tf)
}

func (m *castTransitFrame) ServerTime() *ServerTime {
	return value.Cast[*ServerTime](m.tf)
}

func (m *castTransitFrame) ChannelFrame() ChannelFrame {
	return value.Cast[ChannelFrame](m.tf)
}

func (m *castTransitFrame) JoinChannel() *JoinChannel {
	return value.Cast[*JoinChannel](m.tf)
}

func (m *castTransitFrame) GetChannelMeta() *GetChannelMeta {
	return value.Cast[*GetChannelMeta](m.tf)
}

func (m *castTransitFrame) SetChannelMeta() *SetChannelMeta {
	return value.Cast[*SetChannelMeta](m.tf)
}

func (m *castTransitFrame) ChannelMessage() *ChannelMessage {
	return value.Cast[*ChannelMessage](m.tf)
}

func (m *castTransitFrame) ChannelOwnerMessage() *ChannelOwnerMessage {
	return value.Cast[*ChannelOwnerMessage](m.tf)
}

func (m *castTransitFrame) ChannelCount() *ChannelCount {
	return value.Cast[*ChannelCount](m.tf)
}

func (m *castTransitFrame) LeaveChannel() *LeaveChannel {
	return value.Cast[*LeaveChannel](m.tf)
}

func (m *castTransitFrame) CloseChannel() *CloseChannel {
	return value.Cast[*CloseChannel](m.tf)
}

func (m *castTransitFrame) VoteFrame() VoteFrame {
	return value.Cast[VoteFrame](m.tf)
}

func (m *castTransitFrame) JoinVote() *JoinVote {
	return value.Cast[*JoinVote](m.tf)
}

func (m *castTransitFrame) GetVoteMeta() *GetVoteMeta {
	return value.Cast[*GetVoteMeta](m.tf)
}

func (m *castTransitFrame) SetVoteMeta() *SetVoteMeta {
	return value.Cast[*SetVoteMeta](m.tf)
}

func (m *castTransitFrame) VoteMessage() *VoteMessage {
	return value.Cast[*VoteMessage](m.tf)
}

func (m *castTransitFrame) VoteOwnerMessage() *VoteOwnerMessage {
	return value.Cast[*VoteOwnerMessage](m.tf)
}

func (m *castTransitFrame) VoteCount() *VoteCount {
	return value.Cast[*VoteCount](m.tf)
}

func (m *castTransitFrame) VoteSelect() *VoteSelect {
	return value.Cast[*VoteSelect](m.tf)
}

func (m *castTransitFrame) VoteStatus() *VoteStatus {
	return value.Cast[*VoteStatus](m.tf)
}

func (m *castTransitFrame) LeaveVote() *LeaveVote {
	return value.Cast[*LeaveVote](m.tf)
}

func (m *castTransitFrame) CloseVote() *CloseVote {
	return value.Cast[*CloseVote](m.tf)
}

func (m *castTransitFrame) GetSessionMeta() *GetSessionMeta {
	return value.Cast[*GetSessionMeta](m.tf)
}

func (m *castTransitFrame) SetSessionMeta() *SetSessionMeta {
	return value.Cast[*SetSessionMeta](m.tf)
}

func (m *castTransitFrame) SessionMessage() *SessionMessage {
	return value.Cast[*SessionMessage](m.tf)
}
