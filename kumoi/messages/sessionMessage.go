package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type SessionMessage struct {
	transitFrame
	FromSessionId string
	Message       string
}

func (c *SessionMessage) GetSessionMessage() *SessionMessage {
	return c
}

func (c *SessionMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSessionId = tf.GetSessionMessage().GetFromSession()
	c.Message = tf.GetSessionMessage().GetMessage()
	c.transitFrame.ParseTransitFrame(tf)
	return
}
