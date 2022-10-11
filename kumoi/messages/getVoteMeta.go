package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type GetVoteMeta struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&GetVoteMeta{})
}

func (x *GetVoteMeta) GetVoteId() string {
	return x.BaseTransitFrame().GetGetVoteMeta().GetVoteId()
}

func (x *GetVoteMeta) GetVoteOptions() []*omega.Vote_Option {
	return x.BaseTransitFrame().GetGetVoteMeta().GetVoteOptions()
}

func (x *GetVoteMeta) GetData() map[string]any {
	return base.SafeGetStructMap(x.BaseTransitFrame().GetGetVoteMeta().GetData())
}

func (x *GetVoteMeta) GetName() string {
	return x.BaseTransitFrame().GetGetVoteMeta().GetName()
}

func (x *GetVoteMeta) GetCreatedAt() int64 {
	return x.BaseTransitFrame().GetGetVoteMeta().GetCreatedAt()
}
