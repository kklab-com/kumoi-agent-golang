package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteStatus struct {
	VoteTransitFrame
	Status omega.Vote_Status
}

func (c *VoteStatus) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Status = tf.GetVoteStatus().GetStatus()
	c.VoteTransitFrame.ParseTransitFrame(tf)
	return
}
