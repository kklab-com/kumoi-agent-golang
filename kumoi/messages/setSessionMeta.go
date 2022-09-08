package messages

import (
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SetSessionMeta struct {
	transitFrame
}

func (c *SetSessionMeta) ParseTransitFrame(tf *omega.TransitFrame) {
	c.transitFrame.ParseTransitFrame(tf)
	c.transitFrame.setCast(c)
}
