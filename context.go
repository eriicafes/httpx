package httpx

import (
	"context"
)

type ContextValue[T any] string

func (cv ContextValue[T]) Set(ctx context.Context, val T) context.Context {
	return context.WithValue(ctx, cv, val)
}

func (cv ContextValue[T]) Get(ctx context.Context) (T, bool) {
	val, ok := ctx.Value(cv).(T)
	return val, ok
}
