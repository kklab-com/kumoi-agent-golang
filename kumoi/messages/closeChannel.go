package messages

type CloseChannel struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&CloseChannel{})
}

func (c *CloseChannel) GetChannelId() string {
	return c.BaseTransitFrame().GetCloseChannel().GetChannelId()
}

func (c *CloseChannel) GetOffset() int64 {
	return c.BaseTransitFrame().GetCloseChannel().GetOffset()
}
