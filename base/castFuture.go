package base

import (
	"time"

	concurrent "github.com/kklab-com/goth-concurrent"
	"github.com/kklab-com/goth-kkutil/value"
)

type CastFuture[T any] interface {
	Base() concurrent.Future
	Await() CastFuture[T]
	AwaitTimeout(timeout time.Duration) CastFuture[T]
	Get() T
	GetTimeout(timeout time.Duration) T
	IsDone() bool
	IsSuccess() bool
	IsCancelled() bool
	IsFail() bool
	Error() error
}

func wrapCastFuture[T any](f concurrent.Future) (t CastFuture[T]) {
	return value.Cast[CastFuture[T]](&DefaultCastFuture[T]{bf: f})
}

type DefaultCastFuture[T any] struct {
	bf concurrent.Future
}

func (f *DefaultCastFuture[T]) Base() concurrent.Future {
	return f.bf
}

func (f *DefaultCastFuture[T]) Await() CastFuture[T] {
	f.Base().Await()
	return f
}

func (f *DefaultCastFuture[T]) AwaitTimeout(timeout time.Duration) CastFuture[T] {
	f.Base().AwaitTimeout(timeout)
	return f
}

func (f *DefaultCastFuture[T]) Get() T {
	return value.Cast[T](f.Base().Get())
}

func (f *DefaultCastFuture[T]) GetTimeout(timeout time.Duration) T {
	return value.Cast[T](f.Base().GetTimeout(timeout))
}

func (f *DefaultCastFuture[T]) IsDone() bool {
	return f.Base().IsDone()
}

func (f *DefaultCastFuture[T]) IsSuccess() bool {
	return f.Base().IsSuccess()
}

func (f *DefaultCastFuture[T]) IsCancelled() bool {
	return f.Base().IsCancelled()
}

func (f *DefaultCastFuture[T]) IsFail() bool {
	return f.Base().IsFail()
}

func (f *DefaultCastFuture[T]) Error() error {
	return f.Base().Error()
}
