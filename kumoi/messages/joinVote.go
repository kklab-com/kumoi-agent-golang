package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type JoinVote struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&JoinVote{})
}

func (x *JoinVote) GetName() string {
	return x.BaseTransitFrame().GetJoinVote().GetName()
}

func (x *JoinVote) GetSessionId() string {
	return x.BaseTransitFrame().GetJoinVote().GetSessionId()
}

func (x *JoinVote) GetSessionMetadata() *base.Metadata {
	return x.BaseTransitFrame().GetJoinVote().GetSessionMetadata()
}

func (x *JoinVote) GetVoteId() string {
	return x.BaseTransitFrame().GetJoinVote().GetVoteId()
}

func (x *JoinVote) GetVoteMetadata() *base.Metadata {
	return x.BaseTransitFrame().GetJoinVote().GetVoteMetadata()
}

func (x *JoinVote) GetKey() string {
	return x.BaseTransitFrame().GetJoinVote().GetKey()
}

func (x *JoinVote) GetCreatedAt() int64 {
	return x.BaseTransitFrame().GetJoinVote().GetCreatedAt()
}

func (x *JoinVote) GetVoteOptions() []*omega.Vote_Option {
	return x.BaseTransitFrame().GetJoinVote().GetVoteOptions()
}
