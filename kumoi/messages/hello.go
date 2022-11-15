package messages

type Hello struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&Hello{})
}

func (c *Hello) GetSessionId() string {
	return c.BaseTransitFrame().GetHello().GetSessionId()
}

func (c *Hello) GetSubject() string {
	return c.BaseTransitFrame().GetHello().GetSubject()
}

func (c *Hello) GetSubjectName() string {
	return c.BaseTransitFrame().GetHello().GetSubjectName()
}

func (c *Hello) GetServiceName() string {
	return c.BaseTransitFrame().GetHello().GetServiceName()
}

func (c *Hello) GetServiceVersion() string {
	return c.BaseTransitFrame().GetHello().GetServiceVersion()
}

func (c *Hello) GetServiceNodeName() string {
	return c.BaseTransitFrame().GetHello().GetServiceNodeName()
}
