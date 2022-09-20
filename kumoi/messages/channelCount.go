package messages

type ChannelCount struct {
	TransitFrame
}

func (c *ChannelCount) GetChannelId() string {
	return c.BaseTransitFrame().GetChannelCount().GetChannelId()
}

func (c *ChannelCount) GetOwnerCount() int32 {
	return c.BaseTransitFrame().GetChannelCount().GetOwnerCount()
}

func (c *ChannelCount) GetParticipatorCount() int32 {
	return c.BaseTransitFrame().GetChannelCount().GetParticipatorCount()
}

func (c *ChannelCount) GetCount() int32 {
	return c.BaseTransitFrame().GetChannelCount().GetCount()
}
