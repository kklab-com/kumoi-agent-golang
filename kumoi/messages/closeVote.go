package messages

type CloseVote struct {
	TransitFrame
}

func (c *CloseVote) GetVoteId() string {
	return c.BaseTransitFrame().GetCloseVote().GetVoteId()
}
