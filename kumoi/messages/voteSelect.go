package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteSelect struct {
	VoteTransitFrame
	SessionId    string
	Subject      string
	VoteOptionId string
}

func (c *VoteSelect) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetVoteSelect().GetSessionId()
	c.Subject = tf.GetVoteSelect().GetSubject()
	c.VoteOptionId = tf.GetVoteSelect().GetVoteOptionId()
	c.VoteTransitFrame.ParseTransitFrame(tf)
}
