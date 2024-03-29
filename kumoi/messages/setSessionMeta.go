package messages

import "github.com/kklab-com/kumoi-agent-golang/base"

type SetSessionMeta struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&SetSessionMeta{})
}

func (c *SetSessionMeta) GetSessionId() string {
	return c.BaseTransitFrame().GetSetSessionMeta().GetSessionId()
}

func (c *SetSessionMeta) GetName() string {
	return c.BaseTransitFrame().GetSetSessionMeta().GetName()
}

func (c *SetSessionMeta) GetData() map[string]any {
	return base.SafeGetStructMap(c.BaseTransitFrame().GetSetSessionMeta().GetData())
}
