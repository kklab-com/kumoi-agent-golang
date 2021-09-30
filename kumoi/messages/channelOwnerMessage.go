package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type ChannelOwnerMessage struct {
	ChannelTransitFrame
	FromSession string
	Message     string
	Subject     string
	SubjectName string
}

func (c *ChannelOwnerMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetChannelOwnerMessage().GetFromSession()
	c.Message = tf.GetChannelOwnerMessage().GetMessage()
	c.Subject = tf.GetChannelOwnerMessage().GetSubject()
	c.SubjectName = tf.GetChannelOwnerMessage().GetSubjectName()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
