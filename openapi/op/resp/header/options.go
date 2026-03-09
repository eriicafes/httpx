package header

import (
	"github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI response header.
// May be returned from a type implementing the Header interface:
//
//	func (T) Header() Option
type Option func(*v3.Header, *store.Store)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(h *v3.Header, store *store.Store) {
		for _, opt := range opts {
			opt(h, store)
		}
	}
}

// Reference sets a $ref to a named header component in components/headers.
func Reference(name string) Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Reference = name
	}
}

// Description sets the header description.
func Description(s string) Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Description = s
	}
}

// Required sets whether the header is required. Default: false.
func Required(v bool) Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Required = v
	}
}

// Deprecated marks the header as deprecated. Default: false.
func Deprecated() Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Deprecated = true
	}
}

// AllowEmptyValue allows sending an empty value for the header. Default: false.
func AllowEmptyValue() Option {
	return func(h *v3.Header, _ *store.Store) {
		h.AllowEmptyValue = true
	}
}

// Style sets the serialization style for the header.
func Style(s string) Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Style = s
	}
}

// Explode controls whether arrays and objects generate separate header instances.
// Default: false (simple is the only valid style for headers).
func Explode() Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Explode = true
	}
}

// AllowReserved allows reserved URI characters in the header value. Default: false.
func AllowReserved() Option {
	return func(h *v3.Header, _ *store.Store) {
		h.AllowReserved = true
	}
}

// Example sets an example value for the header.
func Example(v any) Option {
	return func(h *v3.Header, _ *store.Store) {
		h.Example = schema.ToYAMLNode(v)
	}
}

// NamedExample adds a named example to the header.
func NamedExample(name string, value any, opts ...example.Option) Option {
	return func(h *v3.Header, store *store.Store) {
		if h.Examples == nil {
			h.Examples = orderedmap.New[string, *base.Example]()
		}
		h.Examples.Set(name, example.New(store, value, opts...))
	}
}
