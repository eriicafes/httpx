package resp

import (
	"reflect"

	"github.com/eriicafes/httpx/openapi/header"
	"github.com/eriicafes/httpx/openapi/schema"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI response.
type Option func(*v3.Response, *schema.Registry)

// Summary sets a short summary of the response.
func Summary(s string) Option {
	return func(r *v3.Response, _ *schema.Registry) {
		r.Summary = s
	}
}

// Description overrides the response description.
func Description(s string) Option {
	return func(r *v3.Response, _ *schema.Registry) {
		r.Description = s
	}
}

// Header adds a typed header to the response.
func Header[T any](name string, opts ...header.Option) Option {
	return func(r *v3.Response, registry *schema.Registry) {
		if r.Headers == nil {
			r.Headers = orderedmap.New[string, *v3.Header]()
		}
		h := &v3.Header{Schema: schema.ReflectType(reflect.TypeFor[T](), registry)}
		for _, opt := range opts {
			opt(h)
		}
		r.Headers.Set(name, h)
	}
}

// Link adds a design-time link to the response.
func Link(name string, opts ...LinkOption) Option {
	return func(r *v3.Response, _ *schema.Registry) {
		if r.Links == nil {
			r.Links = orderedmap.New[string, *v3.Link]()
		}
		l := &v3.Link{}
		for _, opt := range opts {
			opt(l)
		}
		r.Links.Set(name, l)
	}
}

// ContentType adds a typed content entry to the response.
// Use this to add content types beyond the default application/json set by op.Response[T].
func ContentType[T any](contentType string) Option {
	return func(r *v3.Response, registry *schema.Registry) {
		if r.Content == nil {
			r.Content = orderedmap.New[string, *v3.MediaType]()
		}
		r.Content.Set(contentType, &v3.MediaType{
			Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
		})
	}
}

// Reference sets a $ref to an external response definition.
func Reference(ref string) Option {
	return func(r *v3.Response, _ *schema.Registry) {
		r.Reference = ref
	}
}
