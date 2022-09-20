package messages

type Hello struct {
	TransitFrame
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
