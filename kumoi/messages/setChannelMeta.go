package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SetChannelMeta struct {
	ChannelTransitFrame
	Data *base.Metadata
	Name string
}

func (c *SetChannelMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Data = tf.GetSetChannelMeta().GetData()
	c.Name = tf.GetSetChannelMeta().GetName()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
