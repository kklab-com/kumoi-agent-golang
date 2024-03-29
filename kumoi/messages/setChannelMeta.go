package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SetChannelMeta struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&SetChannelMeta{})
}

func (c *SetChannelMeta) GetChannelId() string {
	return c.BaseTransitFrame().GetSetChannelMeta().GetChannelId()
}

func (c *SetChannelMeta) GetOffset() int64 {
	return c.BaseTransitFrame().GetSetChannelMeta().GetOffset()
}

func (c *SetChannelMeta) GetData() map[string]any {
	return base.SafeGetStructMap(c.BaseTransitFrame().GetSetChannelMeta().GetData())
}

func (c *SetChannelMeta) GetName() string {
	return c.BaseTransitFrame().GetSetChannelMeta().GetName()
}

func (c *SetChannelMeta) GetSkill() *omega.Skill {
	return c.BaseTransitFrame().GetSetChannelMeta().GetSkill()
}
