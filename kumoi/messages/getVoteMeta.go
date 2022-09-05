package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type GetVoteMeta struct {
	VoteTransitFrame
	VoteOptions []VoteOption
	Data        *base.Metadata
	Name        string
	CreatedAt   int64
}

func (c *GetVoteMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.VoteOptions = nil
	for _, vto := range tf.GetVoteCount().GetVoteOptions() {
		c.VoteOptions = append(c.VoteOptions, VoteOption{
			Id:          vto.GetId(),
			Name:        vto.GetName(),
			Count:       vto.GetCount(),
			SubjectList: vto.GetSubjectList(),
		})
	}

	c.Data = tf.GetGetVoteMeta().GetData()
	c.Name = tf.GetGetVoteMeta().GetName()
	c.CreatedAt = tf.GetGetChannelMeta().GetCreatedAt()
	c.VoteTransitFrame.ParseTransitFrame(tf)
}
