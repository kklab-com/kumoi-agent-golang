package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type LeaveVote struct {
	VoteTransitFrame
	SessionId string
}

func (c *LeaveVote) GetLeaveVote() *LeaveVote {
	return c
}

func (c *LeaveVote) ParseTransitFrame(tf *omega.TransitFrame) {
	c.SessionId = tf.GetLeaveVote().GetSessionId()
	c.VoteTransitFrame.ParseTransitFrame(tf)
	return
}
