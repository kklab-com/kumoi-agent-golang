package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type LeaveVote struct {
	voteTransitFrame
	SessionId string
}

func (c *LeaveVote) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetLeaveVote().GetSessionId()
	c.voteTransitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
