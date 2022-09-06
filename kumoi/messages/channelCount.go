package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type ChannelCount struct {
	channelTransitFrame
	OwnerCount        int32
	ParticipatorCount int32
	Count             int32
}

func (c *ChannelCount) ParseTransitFrame(tf *omega.TransitFrame) {
	c.OwnerCount = tf.GetChannelCount().GetOwnerCount()
	c.ParticipatorCount = tf.GetChannelCount().GetParticipatorCount()
	c.Count = tf.GetChannelCount().GetCount()
	c.channelTransitFrame.ParseTransitFrame(tf)
}
