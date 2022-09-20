package kumoi

import (
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
	"github.com/kklab-com/kumoi-agent-golang/base"
	"github.com/kklab-com/kumoi-agent-golang/base/apiresponse"
	"github.com/kklab-com/kumoi-agent-golang/kumoi/messages"
)

type SendFuture[T messages.TransitFrame] interface {
	Base() base.SendFuture
	Await() SendFuture[T]
	AwaitTimeout(timeout time.Duration) SendFuture[T]
	IsDone() bool
	IsSuccess() bool
	IsCancelled() bool
	IsFail() bool
	Error() error
	TransitFrame() (t T)
	AddListener(listener concurrent.FutureListener) SendFuture[T]
}

func wrapSendFuture[T messages.TransitFrame](sf base.SendFuture) (t SendFuture[T]) {
	return value.Cast[SendFuture[T]](&defaultSendFuture[T]{bf: sf})
}

type defaultSendFuture[T messages.TransitFrame] struct {
	bf base.SendFuture
}

func (f *defaultSendFuture[T]) Base() base.SendFuture {
	return f.bf
}

func (f *defaultSendFuture[T]) Await() SendFuture[T] {
	f.Base().Await()
	return f
}

func (f *defaultSendFuture[T]) AwaitTimeout(timeout time.Duration) SendFuture[T] {
	f.Base().AwaitTimeout(timeout)
	return f
}

func (f *defaultSendFuture[T]) IsDone() bool {
	return f.Base().IsDone()
}

func (f *defaultSendFuture[T]) IsSuccess() bool {
	return f.Base().IsSuccess()
}

func (f *defaultSendFuture[T]) IsCancelled() bool {
	return f.Base().IsCancelled()
}

func (f *defaultSendFuture[T]) IsFail() bool {
	return f.Base().IsFail()
}

func (f *defaultSendFuture[T]) Error() error {
	return f.Base().Error()
}

func (f *defaultSendFuture[T]) TransitFrame() (t T) {
	var an any = getParsedTransitFrameFromBaseTransitFrame(f.bf.Get())
	t = value.Cast[T](an)
	return
}

func (f *defaultSendFuture[T]) AddListener(listener concurrent.FutureListener) SendFuture[T] {
	f.Base().AddListener(listener)
	return f
}

type omegaFuture[T any] struct {
	concurrent.CastFuture[T]
	omega *Omega
}

func newOmegaFuture[T any](omega *Omega) omegaFuture[T] {
	return omegaFuture[T]{
		CastFuture: concurrent.NewCastFuture[T](),
		omega:      omega,
	}
}

type CreateChannelFuture struct {
	omegaFuture[*apiresponse.CreateChannel]
}

func (f *CreateChannelFuture) Info() *ChannelInfo {
	if resp := f.Get(); resp != nil {
		return f.omega.Channel(resp.ChannelId)
	}

	return nil
}

func (f *CreateChannelFuture) Join() *Channel {
	if resp := f.Get(); resp != nil {
		return f.Info().Join(resp.OwnerKey)
	}

	return nil
}

type CreateVoteFuture struct {
	omegaFuture[*apiresponse.CreateVote]
}

func (f *CreateVoteFuture) Info() *VoteInfo {
	if resp := f.Get(); resp != nil {
		return f.omega.Vote(resp.VoteId)
	}

	return nil
}

func (f *CreateVoteFuture) Join() *Vote {
	if resp := f.Get(); resp != nil {
		return f.Info().Join(resp.Key)
	}

	return nil
}
