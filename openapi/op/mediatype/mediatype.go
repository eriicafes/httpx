package mediatype

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// MediaType may be implemented to set default media type options on a type.
//
//	func (T) MediaType() mediatype.Option
type MediaType interface {
	MediaType() Option
}

// New builds a media type for T with a reflected schema and any additional options.
func New[T any](store *store.Store, opts ...Option) *v3.MediaType {
	m := &v3.MediaType{Schema: schema.New[T](store)}
	if t, ok := any(new(T)).(MediaType); ok {
		t.MediaType()(m, store)
	}
	for _, opt := range opts {
		opt(m, store)
	}
	return m
}
