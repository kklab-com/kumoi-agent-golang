package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
)

type GetSessionMeta struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&GetSessionMeta{})
}

func (c *GetSessionMeta) GetSessionId() string {
	return c.BaseTransitFrame().GetGetSessionMeta().GetSessionId()
}

func (c *GetSessionMeta) GetSubject() string {
	return c.BaseTransitFrame().GetGetSessionMeta().GetSubject()
}

func (c *GetSessionMeta) GetName() string {
	return c.BaseTransitFrame().GetGetSessionMeta().GetName()
}

func (c *GetSessionMeta) GetData() *base.Metadata {
	return c.BaseTransitFrame().GetGetSessionMeta().GetData()
}
