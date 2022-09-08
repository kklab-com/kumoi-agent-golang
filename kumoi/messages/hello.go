package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type Hello struct {
	transitFrame
	SessionId   string
	Subject     string
	SubjectName string
}

func (c *Hello) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetHello().GetSessionId()
	c.Subject = tf.GetHello().GetSubject()
	c.SubjectName = tf.GetHello().GetSubjectName()
	c.transitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
