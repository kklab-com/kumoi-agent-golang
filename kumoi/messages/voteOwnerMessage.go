package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type VoteOwnerMessage struct {
	VoteTransitFrame
	Message string
}

func (c *VoteOwnerMessage) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Message = tf.GetVoteOwnerMessage().GetMessage()
	c.VoteTransitFrame.ParseTransitFrame(tf)
}
