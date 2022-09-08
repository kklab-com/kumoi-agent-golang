package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelOwnerMessage struct {
	channelTransitFrame
	FromSession string
	Message     string
	Subject     string
	SubjectName string
	Metadata    *base.Metadata
	Skill       *omega.Skill
}

func (c *ChannelOwnerMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetChannelOwnerMessage().GetFromSession()
	c.Message = tf.GetChannelOwnerMessage().GetMessage()
	c.Subject = tf.GetChannelOwnerMessage().GetSubject()
	c.SubjectName = tf.GetChannelOwnerMessage().GetSubjectName()
	c.Metadata = tf.GetChannelOwnerMessage().GetMetadata()
	c.Skill = tf.GetChannelOwnerMessage().GetSkill()
	c.channelTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
