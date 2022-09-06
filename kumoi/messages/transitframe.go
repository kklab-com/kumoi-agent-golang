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

func (m *transitFrame) ProtoMessage() *omega.TransitFrame {
	return m.tf
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
