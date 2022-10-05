package messages

type Broadcast struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&Broadcast{})
}

func (c *Broadcast) GetMessage() string {
	return c.BaseTransitFrame().GetBroadcast().GetMessage()
}
