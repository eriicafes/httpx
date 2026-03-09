package header

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Header is implemented by types that provide default header options.
type Header interface {
	Header() Option
}

func New[T any](store *store.Store, opts ...Option) *v3.Header {
	h := &v3.Header{Schema: schema.New[T](store)}
	if hv, ok := any(new(T)).(Header); ok {
		hv.Header()(h, store)
	}
	for _, opt := range opts {
		opt(h, store)
	}
	// If a reference is set, store the full header in components
	// and return a header $ref object.
	if store != nil && h.Reference != "" {
		h = &v3.Header{Reference: store.SetHeader(h.Reference, h)}
	}
	return h
}
