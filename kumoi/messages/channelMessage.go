package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelMessage struct {
	ChannelTransitFrame
	FromSession   string
	Message       string
	Subject       string
	SubjectName   string
	RoleIndicator omega.Role
}

func (c *ChannelMessage) GetChannelMessage() *ChannelMessage {
	return c
}

func (c *ChannelMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetChannelMessage().GetFromSession()
	c.Message = tf.GetChannelMessage().GetMessage()
	c.Subject = tf.GetChannelMessage().GetSubject()
	c.SubjectName = tf.GetChannelMessage().GetSubjectName()
	c.RoleIndicator = tf.GetChannelMessage().GetRoleIndicator()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
