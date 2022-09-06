package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SetVoteMeta struct {
	voteTransitFrame
	Data *base.Metadata
	Name string
}

func (c *SetVoteMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Data = tf.GetSetVoteMeta().GetData()
	c.Name = tf.GetSetVoteMeta().GetName()
	c.voteTransitFrame.ParseTransitFrame(tf)
}
