package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type CloseVote struct {
	voteTransitFrame
}

func (c *CloseVote) ParseTransitFrame(tf *omega.TransitFrame) {
	c.voteTransitFrame.ParseTransitFrame(tf)
}
