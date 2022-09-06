package messages

import (
	"reflect"

	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelFrame interface {
	GetChannelId() string
	GetOffset() int64
}

type channelTransitFrame struct {
	transitFrame
	channelId string
	offset    int64
}

func (c *channelTransitFrame) GetChannelId() string {
	return c.channelId
}

func (c *channelTransitFrame) GetOffset() int64 {
	return c.offset
}

func (c *channelTransitFrame) ParseTransitFrame(tf *omega.TransitFrame) {
	c.transitFrame.ParseTransitFrame(tf)
	tfd := reflect.ValueOf(tf.GetData())
	if !tfd.IsValid() {
		return
	}

	tfdE := tfd.Elem()
	if tfdE.NumField() == 0 {
		return
	}

	// has ChannelId
	if tfdEv := tfdE.Field(0).Elem().FieldByName("ChannelId"); tfdEv.IsValid() {
		c.channelId = tfdEv.String()
	}

	// has Offset
	if tfdEv := tfdE.Field(0).Elem().FieldByName("Offset"); tfdEv.IsValid() {
		c.offset = tfdEv.Int()
	}
}
