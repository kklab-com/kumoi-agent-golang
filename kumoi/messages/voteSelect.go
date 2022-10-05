package messages

type VoteSelect struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&VoteSelect{})
}

func (x *VoteSelect) GetVoteId() string {
	return x.BaseTransitFrame().GetVoteSelect().GetVoteId()
}

func (x *VoteSelect) GetSessionId() string {
	return x.BaseTransitFrame().GetVoteSelect().GetSessionId()
}

func (x *VoteSelect) GetSubject() string {
	return x.BaseTransitFrame().GetVoteSelect().GetSubject()
}

func (x *VoteSelect) GetVoteOptionId() string {
	return x.BaseTransitFrame().GetVoteSelect().GetVoteOptionId()
}

func (x *VoteSelect) GetDeny() bool {
	return x.BaseTransitFrame().GetVoteSelect().GetDeny()
}
