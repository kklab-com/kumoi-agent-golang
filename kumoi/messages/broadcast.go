package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type Broadcast struct {
	transitFrame
	Message string
}

func (c *Broadcast) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Message = tf.GetBroadcast().GetMessage()
	c.transitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
