package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type GetChannelMeta struct {
	TransitFrame
}

func (c *GetChannelMeta) GetChannelId() string {
	return c.BaseTransitFrame().GetGetChannelMeta().GetChannelId()
}

func (c *GetChannelMeta) GetOffset() int64 {
	return c.BaseTransitFrame().GetGetChannelMeta().GetOffset()
}

func (c *GetChannelMeta) GetName() string {
	return c.BaseTransitFrame().GetGetChannelMeta().GetName()
}

func (c *GetChannelMeta) GetCreatedAt() int64 {
	return c.BaseTransitFrame().GetGetChannelMeta().GetCreatedAt()
}

func (c *GetChannelMeta) GetData() *base.Metadata {
	return c.BaseTransitFrame().GetGetChannelMeta().GetData()
}

func (c *GetChannelMeta) GetSkill() *omega.Skill {
	return c.BaseTransitFrame().GetGetChannelMeta().GetSkill()
}
