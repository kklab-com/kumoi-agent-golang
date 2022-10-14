package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type PlaybackChannelMessage struct {
	TransitFrame
}

func init() {
	registerTransitFrame(&PlaybackChannelMessage{})
}

func (c *PlaybackChannelMessage) GetChannelId() string {
	return c.BaseTransitFrame().GetPlaybackChannelMessage().GetChannelId()
}

func (c *PlaybackChannelMessage) GetNextId() string {
	return c.BaseTransitFrame().GetPlaybackChannelMessage().GetNextId()
}

func (c *PlaybackChannelMessage) GetVolume() omega.Volume {
	return c.BaseTransitFrame().GetPlaybackChannelMessage().GetVolume()
}

func (c *PlaybackChannelMessage) GetCount() int32 {
	return c.BaseTransitFrame().GetPlaybackChannelMessage().GetCount()
}

func (c *PlaybackChannelMessage) GetInverse() bool {
	return c.BaseTransitFrame().GetPlaybackChannelMessage().GetInverse()
}
