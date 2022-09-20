package messages

type Broadcast struct {
	TransitFrame
}

func (c *Broadcast) GetMessage() string {
	return c.BaseTransitFrame().GetBroadcast().GetMessage()
}
