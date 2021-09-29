package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type LeaveChannel struct {
	ChannelTransitFrame
	SessionId string
}

func (c *LeaveChannel) GetLeaveChannel() *LeaveChannel {
	return c
}

func (c *LeaveChannel) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetLeaveChannel().GetSessionId()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
