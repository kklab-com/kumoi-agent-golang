package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type GetChannelMeta struct {
	ChannelTransitFrame
	Data *base.Metadata
	Name string
	CreatedAt int64
}

func (c *GetChannelMeta) GetGetChannelMeta() *GetChannelMeta {
	return c
}

func (c *GetChannelMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Data = tf.GetGetChannelMeta().GetData()
	c.Name = tf.GetGetChannelMeta().GetName()
	c.CreatedAt = tf.GetGetChannelMeta().GetCreatedAt()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
