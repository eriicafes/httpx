package callback

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Option configures an OpenAPI callback object.
// May be returned from a type implementing the Callback interface:
//
//	func (T) Callback() Option
type Option func(*v3.Callback, *store.Store)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(cb *v3.Callback, store *store.Store) {
		for _, opt := range opts {
			opt(cb, store)
		}
	}
}

// Reference sets a $ref to a named callback component in components/callbacks.
func Reference(name string) Option {
	return func(cb *v3.Callback, _ *store.Store) {
		cb.Reference = name
	}
}

// PathItemBuilder is satisfied by pathitem.Path and used to build a *v3.PathItem.
type PathItemBuilder interface {
	Item(*store.Store) *v3.PathItem
}

// Expression adds a runtime expression to path item mapping to the callback.
// Pass a *pathitem.Path as the builder:
//
//	p := pathitem.New(...)
//	p.Operation(http.MethodPost, op.Options(...))
//	callback.Expression("{$url}", p)
func Expression(expr string, p PathItemBuilder) Option {
	return func(cb *v3.Callback, store *store.Store) {
		cb.Expression.Set(expr, p.Item(store))
	}
}
