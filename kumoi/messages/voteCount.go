package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type VoteCount struct {
	TransitFrame
}

func (x *VoteCount) GetVoteId() string {
	return x.BaseTransitFrame().GetVoteCount().GetVoteId()
}

func (x *VoteCount) GetVoteOptions() []*omega.Vote_Option {
	return x.BaseTransitFrame().GetVoteCount().GetVoteOptions()
}
