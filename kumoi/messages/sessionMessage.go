package messages

type SessionMessage struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&SessionMessage{})
}

func (c *SessionMessage) GetFromSession() string {
	return c.BaseTransitFrame().GetSessionMessage().GetFromSession()
}

func (c *SessionMessage) GetToSession() string {
	return c.BaseTransitFrame().GetSessionMessage().GetToSession()
}

func (c *SessionMessage) GetMessage() string {
	return c.BaseTransitFrame().GetSessionMessage().GetMessage()
}

func (c *SessionMessage) GetSequential() bool {
	return c.BaseTransitFrame().GetSessionMessage().GetSequential()
}
