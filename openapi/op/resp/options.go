package resp

import (
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/op/resp/header"
	"github.com/eriicafes/httpx/openapi/op/resp/link"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI response.
// May be returned from a type implementing the Response interface:
//
//	func (T) Response() Option
type Option func(*v3.Response, *store.Store)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(r *v3.Response, store *store.Store) {
		for _, opt := range opts {
			opt(r, store)
		}
	}
}

// Reference sets a $ref to a named response component in components/responses.
func Reference(name string) Option {
	return func(r *v3.Response, _ *store.Store) {
		r.Reference = name
	}
}

// Summary sets a short summary of the response.
func Summary(s string) Option {
	return func(r *v3.Response, _ *store.Store) {
		r.Summary = s
	}
}

// Description overrides the response description.
func Description(s string) Option {
	return func(r *v3.Response, _ *store.Store) {
		r.Description = s
	}
}

// Header adds a typed header to the response.
func Header[T any](name string, opts ...header.Option) Option {
	return func(r *v3.Response, store *store.Store) {
		if r.Headers == nil {
			r.Headers = orderedmap.New[string, *v3.Header]()
		}
		r.Headers.Set(name, header.New[T](store, opts...))
	}
}

// Link adds a design-time link to the response.
func Link(name string, opts ...link.Option) Option {
	return func(r *v3.Response, store *store.Store) {
		if r.Links == nil {
			r.Links = orderedmap.New[string, *v3.Link]()
		}
		r.Links.Set(name, link.New(store, opts...))
	}
}

// ContentType adds a typed content entry to the response.
// Use this to add content types beyond the default application/json set by op.Response[T].
func ContentType[T any](contentType string, opts ...mediatype.Option) Option {
	return func(r *v3.Response, store *store.Store) {
		if r.Content == nil {
			r.Content = orderedmap.New[string, *v3.MediaType]()
		}
		r.Content.Set(contentType, mediatype.New[T](store, opts...))
	}
}
