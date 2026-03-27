package mediatype

import (
	example "github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI media type object.
// May be returned from a type implementing the MediaType interface:
//
//	func (T) MediaType() Option
type Option func(*v3.MediaType, *store.Store)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(m *v3.MediaType, store *store.Store) {
		for _, opt := range opts {
			opt(m, store)
		}
	}
}

// Example sets an example value for the media type.
func Example(v any) Option {
	return func(m *v3.MediaType, _ *store.Store) {
		m.Example = schema.ToYAMLNode(v)
	}
}

// NamedExample adds a named example to the media type.
func NamedExample(name string, value any, opts ...example.Option) Option {
	return func(m *v3.MediaType, store *store.Store) {
		if m.Examples == nil {
			m.Examples = orderedmap.New[string, *base.Example]()
		}
		m.Examples.Set(name, example.New(store, value, opts...))
	}
}

// NamedExampleRef adds a named example $ref to the media type.
func NamedExampleRef(name, ref string) Option {
	return func(m *v3.MediaType, store *store.Store) {
		if m.Examples == nil {
			m.Examples = orderedmap.New[string, *base.Example]()
		}
		m.Examples.Set(name, example.NewRef(ref))
	}
}

// ItemSchema sets the item schema for the media type (OpenAPI 3.2+).
func ItemSchema[T any]() Option {
	return func(m *v3.MediaType, store *store.Store) {
		m.ItemSchema = schema.New[T](store)
	}
}

// Encoding adds a property encoding configuration to the media type.
func Encoding(property string, opts ...EncodingOption) Option {
	return func(m *v3.MediaType, _ *store.Store) {
		if m.Encoding == nil {
			m.Encoding = orderedmap.New[string, *v3.Encoding]()
		}
		e := &v3.Encoding{}
		for _, opt := range opts {
			opt(e)
		}
		m.Encoding.Set(property, e)
	}
}

// ItemEncoding adds a property encoding configuration for the item schema (OpenAPI 3.2+).
func ItemEncoding(property string, opts ...EncodingOption) Option {
	return func(m *v3.MediaType, _ *store.Store) {
		if m.ItemEncoding == nil {
			m.ItemEncoding = orderedmap.New[string, *v3.Encoding]()
		}
		e := &v3.Encoding{}
		for _, opt := range opts {
			opt(e)
		}
		m.ItemEncoding.Set(property, e)
	}
}
