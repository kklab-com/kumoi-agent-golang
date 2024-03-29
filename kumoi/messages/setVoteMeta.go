package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
)

type SetVoteMeta struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&SetVoteMeta{})
}

func (x *SetVoteMeta) GetVoteId() string {
	return x.BaseTransitFrame().GetSetVoteMeta().GetVoteId()
}

func (x *SetVoteMeta) GetData() map[string]any {
	return base.SafeGetStructMap(x.BaseTransitFrame().GetSetVoteMeta().GetData())
}

func (x *SetVoteMeta) GetName() string {
	return x.BaseTransitFrame().GetSetVoteMeta().GetName()
}
