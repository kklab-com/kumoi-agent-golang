package messages

import omega "github.com/kklab-com/kumoi-protobuf-golang"

type CloseChannel struct {
	channelTransitFrame
}

func (c *CloseChannel) ParseTransitFrame(tf *omega.TransitFrame) {
	c.channelTransitFrame.ParseTransitFrame(tf)
}
