package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteStatus struct {
	voteTransitFrame
	Status omega.Vote_Status
}

func (c *VoteStatus) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Status = tf.GetVoteStatus().GetStatus()
	c.voteTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
