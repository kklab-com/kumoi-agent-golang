package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type CloseVote struct {
	VoteTransitFrame
}

func (c *CloseVote) GetCloseVote() *CloseVote {
	return c
}

func (c *CloseVote) ParseTransitFrame(tf *omega.TransitFrame) {
	c.VoteTransitFrame.ParseTransitFrame(tf)
	return
}
