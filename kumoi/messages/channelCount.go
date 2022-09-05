package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type ChannelCount struct {
	ChannelTransitFrame
	OwnerCount        int32
	ParticipatorCount int32
	Count             int32
}

func (c *ChannelCount) ParseTransitFrame(tf *omega.TransitFrame) {
	c.OwnerCount = tf.GetChannelCount().GetOwnerCount()
	c.ParticipatorCount = tf.GetChannelCount().GetParticipatorCount()
	c.Count = tf.GetChannelCount().GetCount()
	c.ChannelTransitFrame.ParseTransitFrame(tf)
}
