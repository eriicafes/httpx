package example

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// Option configures an OpenAPI example object.
type Option func(*base.Example)

// Reference sets a $ref to a named example component in components/examples.
func Reference(ref string) Option {
	return func(e *base.Example) {
		e.Reference = ref
	}
}

// Summary sets a short summary of the example.
func Summary(s string) Option {
	return func(e *base.Example) {
		e.Summary = s
	}
}

// Description sets the long description of the example.
func Description(s string) Option {
	return func(e *base.Example) {
		e.Description = s
	}
}

// Value sets the inline example value.
func Value(v any) Option {
	return func(e *base.Example) {
		e.Value = schema.ToYAMLNode(v)
	}
}

// ExternalValue sets a URL pointing to the literal example value.
// Mutually exclusive with Value.
func ExternalValue(url string) Option {
	return func(e *base.Example) {
		e.ExternalValue = url
	}
}
