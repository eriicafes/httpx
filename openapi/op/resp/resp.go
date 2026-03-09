package resp

import (
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Response may be implemented to set default response options on a type.
//
//	func (T) Response() resp.Option
type Response interface {
	Response() Option
}

func New[T any](store *store.Store, description string, opts ...Option) *v3.Response {
	r := &v3.Response{
		Description: description,
		Content:     orderedmap.New[string, *v3.MediaType](),
	}
	// op.NoContent skips setting a default application/json content entry.
	if _, ok := any(new(T)).(interface{ NoContent() }); !ok {
		r.Content.Set("application/json", mediatype.New[T](store))
	}
	if rv, ok := any(new(T)).(Response); ok {
		rv.Response()(r, store)
	}
	for _, opt := range opts {
		opt(r, store)
	}
	// If a reference is set, store the full response in components
	// and return a response $ref object.
	if store != nil && r.Reference != "" {
		r = &v3.Response{Reference: store.SetResponse(r.Reference, r)}
	}
	return r
}
