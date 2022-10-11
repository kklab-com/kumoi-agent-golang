package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type JoinChannel struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&JoinChannel{})
}

func (c *JoinChannel) GetName() string {
	return c.BaseTransitFrame().GetJoinChannel().GetName()
}

func (c *JoinChannel) GetRole() string {
	return c.BaseTransitFrame().GetJoinChannel().GetRole()
}

func (c *JoinChannel) GetSessionId() string {
	return c.BaseTransitFrame().GetJoinChannel().GetSessionId()
}

func (c *JoinChannel) GetSubject() string {
	return c.BaseTransitFrame().GetJoinChannel().GetSubject()
}

func (c *JoinChannel) GetSubjectName() string {
	return c.BaseTransitFrame().GetJoinChannel().GetSubjectName()
}

func (c *JoinChannel) GetSessionMetadata() map[string]any {
	return base.SafeGetStructMap(c.BaseTransitFrame().GetJoinChannel().GetSessionMetadata())
}

func (c *JoinChannel) GetRoleIndicator() omega.Role {
	return c.BaseTransitFrame().GetJoinChannel().GetRoleIndicator()
}

func (c *JoinChannel) GetSkill() *omega.Skill {
	return c.BaseTransitFrame().GetJoinChannel().GetSkill()
}
