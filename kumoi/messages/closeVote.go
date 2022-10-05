package messages

type CloseVote struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&CloseVote{})
}

func (c *CloseVote) GetVoteId() string {
	return c.BaseTransitFrame().GetCloseVote().GetVoteId()
}
