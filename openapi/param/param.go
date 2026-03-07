package param

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// Option configures an OpenAPI parameter.
type Option func(*v3.Parameter)

// Description sets the parameter description.
func Description(s string) Option {
	return func(p *v3.Parameter) {
		p.Description = s
	}
}

// Required sets whether the parameter is required. Default: false.
func Required(v bool) Option {
	return func(p *v3.Parameter) {
		p.Required = &v
	}
}

// Deprecated marks the parameter as deprecated. Default: false.
func Deprecated() Option {
	return func(p *v3.Parameter) {
		p.Deprecated = true
	}
}

// AllowEmptyValue allows sending an empty value for the parameter. Default: false.
func AllowEmptyValue() Option {
	return func(p *v3.Parameter) {
		p.AllowEmptyValue = true
	}
}

// Style sets the serialization style for the parameter.
// Valid values: matrix, label, form, simple, spaceDelimited, pipeDelimited, deepObject.
func Style(s string) Option {
	return func(p *v3.Parameter) {
		p.Style = s
	}
}

// Explode controls whether arrays and objects generate separate parameter instances.
// Default: true for form style, false for all others.
func Explode(v bool) Option {
	return func(p *v3.Parameter) {
		p.Explode = &v
	}
}

// AllowReserved allows reserved URI characters in query parameter values. Default: false.
func AllowReserved() Option {
	return func(p *v3.Parameter) {
		p.AllowReserved = true
	}
}

// Example sets an example value for the parameter.
func Example(v any) Option {
	return func(p *v3.Parameter) {
		p.Example = toYAMLNode(v)
	}
}

// NamedExample adds a named example to the parameter.
func NamedExample(name string, value any) Option {
	return func(p *v3.Parameter) {
		if p.Examples == nil {
			p.Examples = orderedmap.New[string, *base.Example]()
		}
		p.Examples.Set(name, &base.Example{Value: toYAMLNode(value)})
	}
}

// Reference sets a $ref to an external parameter definition.
func Reference(ref string) Option {
	return func(p *v3.Parameter) {
		p.Reference = ref
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
