package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type JoinChannel struct {
	ChannelTransitFrame
	Name            string
	Role            string
	SessionId       string
	Subject         string
	SubjectName     string
	SessionMetadata *base.Metadata
	RoleIndicator   omega.Role
}

func (c *JoinChannel) GetJoinChannel() *JoinChannel {
	return c
}

func (c *JoinChannel) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Name = tf.GetJoinChannel().GetName()
	c.Role = tf.GetJoinChannel().GetRole()
	c.SessionId = tf.GetJoinChannel().GetSessionId()
	c.Subject = tf.GetJoinChannel().GetSubject()
	c.SubjectName = tf.GetJoinChannel().GetSubjectName()
	c.SessionMetadata = tf.GetJoinChannel().GetSessionMetadata()
	c.RoleIndicator = tf.GetJoinChannel().GetRoleIndicator()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
