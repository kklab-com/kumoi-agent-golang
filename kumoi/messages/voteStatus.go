package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteStatus struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&VoteStatus{})
}

func (x *VoteStatus) GetVoteId() string {
	return x.BaseTransitFrame().GetVoteStatus().GetVoteId()
}

func (x *VoteStatus) GetStatus() omega.Vote_Status {
	return x.BaseTransitFrame().GetVoteStatus().GetStatus()
}
