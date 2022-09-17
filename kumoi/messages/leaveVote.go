package messages

type LeaveVote struct {
	TransitFrame
}

func (c *LeaveVote) GetSessionId() string {
	return c.BaseTransitFrame().GetLeaveVote().GetSessionId()
}

func (c *LeaveVote) GetVoteId() string {
	return c.BaseTransitFrame().GetLeaveVote().GetVoteId()
}
