package body

import (
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// RequestBody may be implemented to set default request body options on a type.
//
//	func (T) RequestBody() body.Option
type RequestBody interface {
	RequestBody() Option
}

// New builds a request body for T.
// Unless T is op.NoContent, it adds an application/json media type by default.
func New[T any](store *store.Store, opts ...Option) *v3.RequestBody {
	required := true
	b := &v3.RequestBody{
		Required: &required,
		Content:  orderedmap.New[string, *v3.MediaType](),
	}
	// op.NoContent skips setting a default application/json content entry.
	if _, ok := any(new(T)).(interface{ NoContent() }); !ok {
		b.Content.Set("application/json", mediatype.New[T](store))
	}
	if t, ok := any(new(T)).(RequestBody); ok {
		t.RequestBody()(b, store)
	}
	if b.Reference == "" {
		// Skip inline options if reference is set on type
		for _, opt := range opts {
			opt(b, store)
		}
	}
	// If a reference is set, store the full request body in components
	// and return a request body $ref object.
	if store != nil && b.Reference != "" {
		b = store.SetRequestBody(b.Reference, b)
	}
	return b
}
