package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type ReplayChannelMessage struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&ReplayChannelMessage{})
}

func (c *ReplayChannelMessage) GetChannelId() string {
	return c.BaseTransitFrame().GetReplayChannelMessage().GetChannelId()
}

func (c *ReplayChannelMessage) GetNextId() string {
	return c.BaseTransitFrame().GetReplayChannelMessage().GetNextId()
}

func (c *ReplayChannelMessage) GetVolume() omega.Volume {
	return c.BaseTransitFrame().GetReplayChannelMessage().GetVolume()
}

func (c *ReplayChannelMessage) GetCount() int32 {
	return c.BaseTransitFrame().GetReplayChannelMessage().GetCount()
}

func (c *ReplayChannelMessage) GetInverse() bool {
	return c.BaseTransitFrame().GetReplayChannelMessage().GetInverse()
}
