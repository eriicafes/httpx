package header

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Header may be implemented to set default header options on a type.
//
//	func (T) Header() header.Option
type Header interface {
	Header() Option
}

// New builds a response header for T.
func New[T any](store *store.Store, opts ...Option) *v3.Header {
	h := &v3.Header{Schema: schema.New[T](store)}
	if t, ok := any(new(T)).(Header); ok {
		t.Header()(h, store)
	}
	if h.Reference == "" {
		// Skip inline options if reference is set on type
		for _, opt := range opts {
			opt(h, store)
		}
	}
	// If a reference is set, store the full header in components
	// and return a header $ref object.
	if store != nil && h.Reference != "" {
		h = store.SetHeader(h.Reference, h)
	}
	return h
}
