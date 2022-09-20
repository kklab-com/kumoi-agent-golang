package messages

type LeaveChannel struct {
	TransitFrame
}

func (c *LeaveChannel) GetSessionId() string {
	return c.BaseTransitFrame().GetLeaveChannel().GetSessionId()
}

func (c *LeaveChannel) GetChannelId() string {
	return c.BaseTransitFrame().GetLeaveChannel().GetChannelId()
}

func (c *LeaveChannel) GetOffset() int64 {
	return c.BaseTransitFrame().GetLeaveChannel().GetOffset()
}
