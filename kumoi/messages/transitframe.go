package messages

import (
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
	ProtoMessage() *omega.TransitFrame
}

type TransitFrameParsable interface {
	ParseTransitFrame(tf *omega.TransitFrame)
}

type transitFrame struct {
	TransitId        uint64
	Class            omega.TransitFrame_FrameClass
	Version          omega.TransitFrame_Version
	Timestamp        int64
	Err              *omega.Error
	MessageId        string
	RefererMessageId string
	tf               *omega.TransitFrame
}

func (m *transitFrame) GetTransitId() uint64 {
	return m.TransitId
}

func (m *transitFrame) GetClass() omega.TransitFrame_FrameClass {
	return m.Class
}

func (m *transitFrame) GetVersion() omega.TransitFrame_Version {
	return m.Version
}

func (m *transitFrame) GetTimestamp() int64 {
	return m.Timestamp
}

func (m *transitFrame) GetErr() *omega.Error {
	return m.Err
}

func (m *transitFrame) GetMessageId() string {
	return m.MessageId
}

func (m *transitFrame) GetRefererMessageId() string {
	return m.RefererMessageId
}

func (m *transitFrame) ProtoMessage() *omega.TransitFrame {
	return m.tf
}

func (m *transitFrame) Error() string {
	if m.Err == nil {
		return ""
	}

	return value.JsonMarshal(m.Err)
}

func (m *transitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	m.TransitId = tf.GetTransitId()
	m.Class = tf.GetClass()
	m.Version = tf.GetVersion()
	m.Timestamp = tf.GetTimestamp()
	m.MessageId = TransitMessageIdCodec.EncodeToString(tf.GetMessageId())
	m.RefererMessageId = TransitMessageIdCodec.EncodeToString(tf.GetRefererMessageId())
	m.tf = tf
}
