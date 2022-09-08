package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SetChannelMeta struct {
	channelTransitFrame
	Data  *base.Metadata
	Name  string
	Skill *omega.Skill
}

func (c *SetChannelMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Data = tf.GetSetChannelMeta().GetData()
	c.Name = tf.GetSetChannelMeta().GetName()
	c.Skill = tf.GetSetChannelMeta().GetSkill()
	c.channelTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
