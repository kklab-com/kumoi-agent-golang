package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type LeaveChannel struct {
	channelTransitFrame
	SessionId string
}

func (c *LeaveChannel) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetLeaveChannel().GetSessionId()
	c.channelTransitFrame.ParseTransitFrame(tf)
}
