package body

import (
	"reflect"

	"github.com/eriicafes/httpx/openapi/schema"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI request body.
type Option func(*v3.RequestBody, *schema.Registry)

// Description sets the request body description.
func Description(s string) Option {
	return func(b *v3.RequestBody, _ *schema.Registry) {
		b.Description = s
	}
}

// Required marks the request body as required. Default: false.
func Required() Option {
	required := true
	return func(b *v3.RequestBody, _ *schema.Registry) {
		b.Required = &required
	}
}

// ContentType adds a typed content entry to the request body.
// Use this to add content types beyond the default application/json set by op.RequestBody[T].
func ContentType[T any](contentType string) Option {
	return func(b *v3.RequestBody, registry *schema.Registry) {
		if b.Content == nil {
			b.Content = orderedmap.New[string, *v3.MediaType]()
		}
		b.Content.Set(contentType, &v3.MediaType{
			Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
		})
	}
}

// Reference sets a $ref to an external request body definition.
func Reference(ref string) Option {
	return func(b *v3.RequestBody, _ *schema.Registry) {
		b.Reference = ref
	}
}
