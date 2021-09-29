package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type Time struct {
	transitFrame
	UnixNano      int64
	UnixTimestamp int64
}

func (c *Time) ParseTransitFrame(tf *omega.TransitFrame) {
	c.UnixNano = tf.GetServerTime().GetUnixNano()
	c.UnixTimestamp = tf.GetServerTime().GetUnixTimestamp()
	c.transitFrame.ParseTransitFrame(tf)
	return
}
