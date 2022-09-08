package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteMessage struct {
	voteTransitFrame
	FromSession   string
	Message       string
	Subject       string
	RoleIndicator omega.Role
}

func (c *VoteMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.FromSession = tf.GetVoteMessage().GetFromSession()
	c.Message = tf.GetVoteMessage().GetMessage()
	c.Subject = tf.GetVoteMessage().GetSubject()
	c.RoleIndicator = tf.GetVoteMessage().GetRoleIndicator()
	c.voteTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
