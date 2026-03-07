package header

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// Option configures an OpenAPI response header.
type Option func(*v3.Header)

// Description sets the header description.
func Description(s string) Option {
	return func(h *v3.Header) {
		h.Description = s
	}
}

// Required marks the header as required. Default: false.
func Required() Option {
	return func(h *v3.Header) {
		h.Required = true
	}
}

// Deprecated marks the header as deprecated. Default: false.
func Deprecated() Option {
	return func(h *v3.Header) {
		h.Deprecated = true
	}
}

// AllowEmptyValue allows sending an empty value for the header. Default: false.
func AllowEmptyValue() Option {
	return func(h *v3.Header) {
		h.AllowEmptyValue = true
	}
}

// Style sets the serialization style for the header.
func Style(s string) Option {
	return func(h *v3.Header) {
		h.Style = s
	}
}

// Explode controls whether arrays and objects generate separate header instances.
// Default: false (simple is the only valid style for headers).
func Explode() Option {
	return func(h *v3.Header) {
		h.Explode = true
	}
}

// AllowReserved allows reserved URI characters in the header value. Default: false.
func AllowReserved() Option {
	return func(h *v3.Header) {
		h.AllowReserved = true
	}
}

// Example sets an example value for the header.
func Example(v any) Option {
	return func(h *v3.Header) {
		h.Example = toYAMLNode(v)
	}
}

// NamedExample adds a named example to the header.
func NamedExample(name string, value any) Option {
	return func(h *v3.Header) {
		if h.Examples == nil {
			h.Examples = orderedmap.New[string, *base.Example]()
		}
		h.Examples.Set(name, &base.Example{Value: toYAMLNode(value)})
	}
}

// Reference sets a $ref to an external header definition.
func Reference(ref string) Option {
	return func(h *v3.Header) {
		h.Reference = ref
	}
}

func toYAMLNode(v any) *yaml.Node {
	node := &yaml.Node{}
	b, _ := yaml.Marshal(v)
	_ = yaml.Unmarshal(b, node)
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}
	return node
}
