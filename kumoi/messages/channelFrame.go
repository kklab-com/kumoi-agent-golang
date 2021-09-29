package messages

import (
	"reflect"

	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelFrame interface {
	ChannelId() string
	Offset() int64
}

type ChannelTransitFrame struct {
	transitFrame
	channelId string
	offset    int64
}

func (c *ChannelTransitFrame) ChannelId() string {
	return c.channelId
}

func (c *ChannelTransitFrame) Offset() int64 {
	return c.offset
}

func (c *ChannelTransitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	c.transitFrame.ParseTransitFrame(tf)
	tfd := reflect.ValueOf(tf.GetData())
	if !tfd.IsValid() {
		return
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return
	}

	// has channelId
	if tfdEv := tfdE.Field(0).Elem().FieldByName("ChannelId"); tfdEv.IsValid() {
		c.channelId = tfdEv.String()
	}

	// has offset
	if tfdEv := tfdE.Field(0).Elem().FieldByName("Offset"); tfdEv.IsValid() {
		c.offset = tfdEv.Int()
	}
}
