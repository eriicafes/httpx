// Package contextkey provides type-safe context key management using Go generics.
// It wraps the standard context.Context value storage with compile-time type safety,
// eliminating the need for type assertions and reducing runtime errors.
package contextkey

import "context"

// Key is a type-safe interface for storing and retrieving values in a context.
// It provides compile-time type safety for context values through Go generics.
type Key[T any] interface {
	// Get retrieves the value associated with this key from the context.
	// Returns the value and true if found, or the zero value and false if not found.
	Get(context.Context) (T, bool)

	// Set stores a value in the context and returns a new context with the value set.
	Set(context.Context, T) context.Context

	// Delete removes the value associated with this key from the context.
	// Returns a new context with the value removed.
	Delete(context.Context) context.Context

	// Take retrieves and removes the value in a single operation.
	// Returns the new context, the value, and whether the value was found.
	// If the value doesn't exist, returns the original context and the zero value.
	Take(context.Context) (context.Context, T, bool)

	// Update modifies the existing value using the provided update function.
	// The update function receives the previous value and whether it existed.
	// Returns a new context with the updated value.
	Update(ctx context.Context, update func(prev T, hasPrev bool) T) context.Context
}

// New creates a new type-safe context key with the given name.
// The name should be unique to avoid collisions with other context keys.
func New[T any](name string) Key[T] {
	return key[T](name)
}

type key[T any] string

func (k key[T]) Get(ctx context.Context) (T, bool) {
	value, ok := ctx.Value(k).(T)
	return value, ok
}

func (k key[T]) Set(ctx context.Context, value T) context.Context {
	return context.WithValue(ctx, k, value)
}

func (k key[T]) Delete(ctx context.Context) context.Context {
	return context.WithValue(ctx, k, nil)
}

func (k key[T]) Take(ctx context.Context) (context.Context, T, bool) {
	value, ok := ctx.Value(k).(T)
	if !ok {
		return ctx, value, ok
	}
	return context.WithValue(ctx, k, nil), value, ok
}

func (k key[T]) Update(ctx context.Context, update func(prev T, hasPrev bool) T) context.Context {
	value, ok := ctx.Value(k).(T)
	return context.WithValue(ctx, k, update(value, ok))
}
