package param

import (
	"github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI parameter.
type Option func(*v3.Parameter, *store.Store)

// Reference sets a $ref to a named parameter component in components/parameters.
func Reference(name string) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Reference = name
	}
}

// Description sets the parameter description.
func Description(s string) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Description = s
	}
}

// Required sets whether the parameter is required. Default: false.
func Required(v bool) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Required = &v
	}
}

// Deprecated marks the parameter as deprecated. Default: false.
func Deprecated() Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Deprecated = true
	}
}

// AllowEmptyValue allows sending an empty value for the parameter. Default: false.
func AllowEmptyValue() Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.AllowEmptyValue = true
	}
}

// Style sets the serialization style for the parameter.
// Valid values: matrix, label, form, simple, spaceDelimited, pipeDelimited, deepObject.
func Style(s string) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Style = s
	}
}

// Explode controls whether arrays and objects generate separate parameter instances.
// Default: true for form style, false for all others.
func Explode(v bool) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Explode = &v
	}
}

// AllowReserved allows reserved URI characters in query parameter values. Default: false.
func AllowReserved() Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.AllowReserved = true
	}
}

// Example sets an example value for the parameter.
func Example(v any) Option {
	return func(p *v3.Parameter, _ *store.Store) {
		p.Example = schema.ToYAMLNode(v)
	}
}

// NamedExample adds a named example to the parameter.
func NamedExample(name string, opts ...example.Option) Option {
	return func(p *v3.Parameter, store *store.Store) {
		if p.Examples == nil {
			p.Examples = orderedmap.New[string, *base.Example]()
		}
		p.Examples.Set(name, example.New(store, opts...))
	}
}
