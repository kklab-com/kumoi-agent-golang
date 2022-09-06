package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type VoteCount struct {
	voteTransitFrame
	VoteOptions []VoteOption
}

func (c *VoteCount) ParseTransitFrame(tf *omega.TransitFrame) {
	c.VoteOptions = nil
	for _, vto := range tf.GetVoteCount().GetVoteOptions() {
		c.VoteOptions = append(c.VoteOptions, VoteOption{
			Id:          vto.GetId(),
			Name:        vto.GetName(),
			Count:       vto.GetCount(),
			SubjectList: vto.GetSubjectList(),
		})
	}

	c.voteTransitFrame.ParseTransitFrame(tf)
}
