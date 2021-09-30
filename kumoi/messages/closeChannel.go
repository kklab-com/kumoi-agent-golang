package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type CloseChannel struct {
	ChannelTransitFrame
}

func (c *CloseChannel) ParseTransitFrame(tf *omega.TransitFrame) {
	c.ChannelTransitFrame.ParseTransitFrame(tf)
	return
}
