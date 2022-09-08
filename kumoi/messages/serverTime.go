package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type ServerTime struct {
	transitFrame
	UnixNano      int64
	UnixTimestamp int64
}

func (c *ServerTime) ParseTransitFrame(tf *omega.TransitFrame) {
	c.UnixNano = tf.GetServerTime().GetUnixNano()
	c.UnixTimestamp = tf.GetServerTime().GetUnixTimestamp()
	c.transitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
