package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelOwnerMessage struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&ChannelOwnerMessage{})
}

func (c *ChannelOwnerMessage) GetChannelId() string {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetChannelId()
}

func (c *ChannelOwnerMessage) GetOffset() int64 {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetOffset()
}

func (c *ChannelOwnerMessage) GetFromSession() string {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetFromSession()
}

func (c *ChannelOwnerMessage) GetMessage() string {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetMessage()
}

func (c *ChannelOwnerMessage) GetSubject() string {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetSubject()
}

func (c *ChannelOwnerMessage) GetSubjectName() string {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetSubjectName()
}

func (c *ChannelOwnerMessage) GetMetadata() map[string]any {
	return base.SafeGetStructMap(c.BaseTransitFrame().GetChannelOwnerMessage().GetMetadata())
}

func (c *ChannelOwnerMessage) GetSkill() *omega.Skill {
	return c.BaseTransitFrame().GetChannelOwnerMessage().GetSkill()
}
