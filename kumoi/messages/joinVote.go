package messages

import (
	"github.com/kklab-com/kumoi-agent-golang/base"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type JoinVote struct {
	VoteTransitFrame
	Name            string
	SessionId       string
	SessionMetadata *base.Metadata
}

func (c *JoinVote) GetJoinVote() *JoinVote {
	return c
}

func (c *JoinVote) ParseTransitFrame(tf *omega.TransitFrame) {
	c.Name = tf.GetJoinVote().GetName()
	c.SessionId = tf.GetJoinVote().GetSessionId()
	c.SessionMetadata = tf.GetJoinVote().GetSessionMetadata()
	c.VoteTransitFrame.ParseTransitFrame(tf)
	return
}
