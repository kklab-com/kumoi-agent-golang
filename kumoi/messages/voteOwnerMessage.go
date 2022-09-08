package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type VoteOwnerMessage struct {
	voteTransitFrame
	Message string
}

func (c *VoteOwnerMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Message = tf.GetVoteOwnerMessage().GetMessage()
	c.voteTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
