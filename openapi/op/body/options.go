package body

import (
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI request body.
type Option func(*v3.RequestBody, *store.Store)

// Reference sets a $ref to a named request body component in components/requestBodies.
func Reference(name string) Option {
	return func(b *v3.RequestBody, _ *store.Store) {
		b.Reference = name
	}
}

// Description sets the request body description.
func Description(s string) Option {
	return func(b *v3.RequestBody, _ *store.Store) {
		b.Description = s
	}
}

// Required sets whether the request body is required. Default: true.
func Required(v bool) Option {
	return func(b *v3.RequestBody, _ *store.Store) {
		b.Required = &v
	}
}

// ContentType adds a typed content entry to the request body.
// Use this to add content types beyond the default application/json set by op.RequestBody[T].
func ContentType[T any](contentType string, opts ...mediatype.Option) Option {
	return func(b *v3.RequestBody, store *store.Store) {
		if b.Content == nil {
			b.Content = orderedmap.New[string, *v3.MediaType]()
		}
		b.Content.Set(contentType, mediatype.New[T](store, opts...))
	}
}
