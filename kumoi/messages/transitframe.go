package messages

import (
	"reflect"

	"github.com/kklab-com/goth-kkutil/value"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

var TransitMessageIdCodec = &omega.TransitMessageIDCodec{}

var transitFrameMaps = map[string]reflect.Type{}

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

func registerTransitFrame(tf TransitFrame) {
	transitFrameMaps[reflect.TypeOf(tf).Elem().Name()] = reflect.TypeOf(tf).Elem()
}

func WrapTransitFrame(btf *omega.TransitFrame) (msg TransitFrame) {
	if v, f := transitFrameMaps[reflect.TypeOf(btf.GetData()).Elem().Field(0).Name]; f {
		tf := reflect.New(v)
		tf.Elem().FieldByName("TransitFrame").Set(reflect.ValueOf(&transitFrame{tf: btf}))
		msg = tf.Interface().(TransitFrame)
		return
	}

	return nil
}

type transitFrame struct {
	tf *omega.TransitFrame
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

	return tfdE.Type().Field(0).Name
}

func (m *transitFrame) Cast() CastTransitFrame {
	return &castTransitFrame{tf: m}
}

func (m *transitFrame) Error() string {
	if err := m.GetErr(); err != nil {
		return value.JsonMarshal(m.GetErr())
	}

	return ""
}

type CastTransitFrame interface {
	Broadcast() *Broadcast
	Hello() *Hello
	ServerTime() *ServerTime
	JoinChannel() *JoinChannel
	GetChannelMeta() *GetChannelMeta
	SetChannelMeta() *SetChannelMeta
	ChannelMessage() *ChannelMessage
	ChannelOwnerMessage() *ChannelOwnerMessage
	ChannelCount() *ChannelCount
	PlaybackChannelMessage() *PlaybackChannelMessage
	ReplayChannelMessage() *ReplayChannelMessage
	LeaveChannel() *LeaveChannel
	CloseChannel() *CloseChannel
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

func (m *castTransitFrame) Broadcast() *Broadcast {
	return &Broadcast{TransitFrame: m.tf}
}

func (m *castTransitFrame) Hello() *Hello {
	return &Hello{TransitFrame: m.tf}
}

func (m *castTransitFrame) ServerTime() *ServerTime {
	return &ServerTime{TransitFrame: m.tf}
}

func (m *castTransitFrame) JoinChannel() *JoinChannel {
	return &JoinChannel{TransitFrame: m.tf}
}

func (m *castTransitFrame) GetChannelMeta() *GetChannelMeta {
	return &GetChannelMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) SetChannelMeta() *SetChannelMeta {
	return &SetChannelMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) ChannelMessage() *ChannelMessage {
	return &ChannelMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) ChannelOwnerMessage() *ChannelOwnerMessage {
	return &ChannelOwnerMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) ChannelCount() *ChannelCount {
	return &ChannelCount{TransitFrame: m.tf}
}

func (m *castTransitFrame) PlaybackChannelMessage() *PlaybackChannelMessage {
	return &PlaybackChannelMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) ReplayChannelMessage() *ReplayChannelMessage {
	return &ReplayChannelMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) LeaveChannel() *LeaveChannel {
	return &LeaveChannel{TransitFrame: m.tf}
}

func (m *castTransitFrame) CloseChannel() *CloseChannel {
	return &CloseChannel{TransitFrame: m.tf}
}

func (m *castTransitFrame) JoinVote() *JoinVote {
	return &JoinVote{TransitFrame: m.tf}
}

func (m *castTransitFrame) GetVoteMeta() *GetVoteMeta {
	return &GetVoteMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) SetVoteMeta() *SetVoteMeta {
	return &SetVoteMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) VoteMessage() *VoteMessage {
	return &VoteMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) VoteOwnerMessage() *VoteOwnerMessage {
	return &VoteOwnerMessage{TransitFrame: m.tf}
}

func (m *castTransitFrame) VoteCount() *VoteCount {
	return &VoteCount{TransitFrame: m.tf}
}

func (m *castTransitFrame) VoteSelect() *VoteSelect {
	return &VoteSelect{TransitFrame: m.tf}
}

func (m *castTransitFrame) VoteStatus() *VoteStatus {
	return &VoteStatus{TransitFrame: m.tf}
}

func (m *castTransitFrame) LeaveVote() *LeaveVote {
	return &LeaveVote{TransitFrame: m.tf}
}

func (m *castTransitFrame) CloseVote() *CloseVote {
	return &CloseVote{TransitFrame: m.tf}
}

func (m *castTransitFrame) GetSessionMeta() *GetSessionMeta {
	return &GetSessionMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) SetSessionMeta() *SetSessionMeta {
	return &SetSessionMeta{TransitFrame: m.tf}
}

func (m *castTransitFrame) SessionMessage() *SessionMessage {
	return &SessionMessage{TransitFrame: m.tf}
}
