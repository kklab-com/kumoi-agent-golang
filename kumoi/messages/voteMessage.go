package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteMessage struct {
	VoteTransitFrame
	FromSession   string
	Message       string
	Subject       string
	RoleIndicator omega.Role
}

func (c *VoteMessage) GetVoteMessage() *VoteMessage {
	return c
}

func (c *VoteMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetVoteMessage().GetFromSession()
	c.Message = tf.GetVoteMessage().GetMessage()
	c.Subject = tf.GetVoteMessage().GetSubject()
	c.RoleIndicator = tf.GetVoteMessage().GetRoleIndicator()
	c.VoteTransitFrame.ParseTransitFrame(tf)
	return
}
