package messages

type ServerTime struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&ServerTime{})
}

func (c *ServerTime) GetUnixNano() int64 {
	return c.BaseTransitFrame().GetServerTime().GetUnixNano()
}

func (c *ServerTime) GetUnixTimestamp() int64 {
	return c.BaseTransitFrame().GetServerTime().GetUnixTimestamp()
}
