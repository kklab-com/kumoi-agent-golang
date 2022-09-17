package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type ChannelMessage struct {
	TransitFrame
}

func (c *ChannelMessage) GetChannelId() string {
	return c.BaseTransitFrame().GetChannelMessage().GetChannelId()
}

func (c *ChannelMessage) GetOffset() int64 {
	return c.BaseTransitFrame().GetChannelMessage().GetOffset()
}

func (c *ChannelMessage) GetFromSession() string {
	return c.BaseTransitFrame().GetChannelMessage().GetFromSession()
}

func (c *ChannelMessage) GetMessage() string {
	return c.BaseTransitFrame().GetChannelMessage().GetMessage()
}

func (c *ChannelMessage) GetSubject() string {
	return c.BaseTransitFrame().GetChannelMessage().GetSubject()
}

func (c *ChannelMessage) GetSubjectName() string {
	return c.BaseTransitFrame().GetChannelMessage().GetSubjectName()
}

func (c *ChannelMessage) GetRoleIndicator() omega.Role {
	return c.BaseTransitFrame().GetChannelMessage().GetRoleIndicator()
}

func (c *ChannelMessage) GetMetadata() *base.Metadata {
	return c.BaseTransitFrame().GetChannelMessage().GetMetadata()
}

func (c *ChannelMessage) GetSkill() *omega.Skill {
	return c.BaseTransitFrame().GetChannelMessage().GetSkill()
}
