package mediatype

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// MediaType is implemented by types that provide default media type options.
type MediaType interface {
	MediaType() Option
}

func New[T any](store *store.Store, opts ...Option) *v3.MediaType {
	m := &v3.MediaType{Schema: schema.New[T](store)}
	if mv, ok := any(new(T)).(MediaType); ok {
		mv.MediaType()(m, store)
	}
	for _, opt := range opts {
		opt(m, store)
	}
	return m
}
