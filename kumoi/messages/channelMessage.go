package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelMessage struct {
	channelTransitFrame
	FromSession   string
	Message       string
	Subject       string
	SubjectName   string
	RoleIndicator omega.Role
	Metadata      *base.Metadata
	Skill         *omega.Skill
}

func (c *ChannelMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetChannelMessage().GetFromSession()
	c.Message = tf.GetChannelMessage().GetMessage()
	c.Subject = tf.GetChannelMessage().GetSubject()
	c.SubjectName = tf.GetChannelMessage().GetSubjectName()
	c.RoleIndicator = tf.GetChannelMessage().GetRoleIndicator()
	c.Metadata = tf.GetChannelMessage().GetMetadata()
	c.Skill = tf.GetChannelMessage().GetSkill()
	c.channelTransitFrame.ParseTransitFrame(tf)
}
