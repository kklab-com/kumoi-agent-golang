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

func (m *transitFrame) Error() string {
	if err := m.GetErr(); err != nil {
		return value.JsonMarshal(m.GetErr())
	}

	return ""
}
