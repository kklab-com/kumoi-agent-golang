package messages

import "github.com/kklab-com/kumoi-agent-golang/base"

type SetSessionMeta struct {
	TransitFrame
}

func (c *SetSessionMeta) GetSessionId() string {
	return c.BaseTransitFrame().GetSetSessionMeta().GetSessionId()
}

func (c *SetSessionMeta) GetName() string {
	return c.BaseTransitFrame().GetSetSessionMeta().GetName()
}

func (c *SetSessionMeta) GetData() *base.Metadata {
	return c.BaseTransitFrame().GetSetSessionMeta().GetData()
}
