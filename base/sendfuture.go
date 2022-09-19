package base

import (
	concurrent "github.com/kklab-com/goth-concurrent"
	omega "github.com/kklab-com/kumoi-protobuf-golang"
)

type SendFuture interface {
	concurrent.CastFuture[*omega.TransitFrame]
	SentTransitFrame() *omega.TransitFrame
}

type DefaultSendFuture struct {
	concurrent.CastFuture[*omega.TransitFrame]
	tf *omega.TransitFrame
}

func (f *DefaultSendFuture) SentTransitFrame() *omega.TransitFrame {
	return f.tf
}
