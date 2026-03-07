package schema

import (
	"reflect"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

// SchemaOption allows a type to customise its generated JSON Schema.
type SchemaOption interface {
	Schema() Option
}

// ReflectType builds a SchemaProxy from a reflect.Type, applying SchemaOption constraints
// if the type implements the interface. Pass a non-nil registry to enable component
// auto-registration for types that declare a Reference via schema.Ref.
func ReflectType(t reflect.Type, registry *Registry) *base.SchemaProxy {
	return reflectType(t, make(map[reflect.Type]bool), registry)
}

func reflectType(t reflect.Type, visited map[reflect.Type]bool, registry *Registry) *base.SchemaProxy {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}

	return registry.Get(t, func(rf *RegistryField) *base.SchemaProxy {
		// TODO: t can no longer be pointer type so nullable check never runs
		s := schemaFromType(t, visited, registry)
		proxy := base.CreateSchemaProxy(s)

		// Interface types cannot be allocated; skip the SchemaOption check for them.
		if t.Kind() != reflect.Interface {
			v := reflect.New(t)
			if sc, ok := v.Elem().Interface().(SchemaOption); ok {
				sc.Schema()(s, rf)
			} else if sc, ok := v.Interface().(SchemaOption); ok {
				sc.Schema()(s, rf)
			}
		}

		return proxy
	})
}

// SchemaFromType builds a base.Schema from a reflect.Type without registry support.
func SchemaFromType(t reflect.Type) *base.Schema {
	return schemaFromType(t, make(map[reflect.Type]bool), nil)
}

func schemaFromType(t reflect.Type, visited map[reflect.Type]bool, registry *Registry) *base.Schema {
	if t.Kind() == reflect.Pointer {
		s := schemaFromType(t.Elem(), visited, registry)
		nullable := true
		s.Nullable = &nullable
		return s
	}

	switch t.Kind() {
	case reflect.String:
		return &base.Schema{Type: []string{"string"}}
	case reflect.Bool:
		return &base.Schema{Type: []string{"boolean"}}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return &base.Schema{Type: []string{"integer"}}
	case reflect.Float32, reflect.Float64:
		return &base.Schema{Type: []string{"number"}}
	case reflect.Slice, reflect.Array:
		return &base.Schema{
			Type: []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: reflectType(t.Elem(), visited, registry),
			},
		}
	case reflect.Map:
		return &base.Schema{
			Type: []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: reflectType(t.Elem(), visited, registry),
			},
		}
	case reflect.Struct:
		// union types expand to oneOf / tagged oneOf with discriminator
		if info := unionFromType(t); info != nil {
			s := &base.Schema{OneOf: info.schemas}
			if info.discriminator != nil {
				s.Discriminator = info.discriminator
			}
			return s
		}

		// special case time.Time
		if t.PkgPath() == "time" && t.Name() == "Time" {
			return &base.Schema{
				Type:   []string{"string"},
				Format: "date-time",
			}
		}

		// Guard against self-referential types (e.g. type Node struct { Next *Node }).
		if visited[t] {
			return &base.Schema{}
		}
		visited[t] = true

		props := orderedmap.New[string, *base.SchemaProxy]()
		required := []string{}

		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			if !f.IsExported() {
				continue
			}
			name := f.Name
			if tag, ok := f.Tag.Lookup("json"); ok {
				tagName, _, _ := strings.Cut(tag, ",")
				if tagName == "-" {
					continue
				}
				if tagName != "" {
					name = tagName
				}
			}
			props.Set(name, reflectType(f.Type, visited, registry))
			if f.Type.Kind() != reflect.Pointer {
				required = append(required, name)
			}
		}
		return &base.Schema{
			Type:       []string{"object"},
			Properties: props,
			Required:   required,
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 1,
				B: false,
			},
		}
	default:
		// Unsupported types (Interface, Chan, Func, Complex, etc.) fall back to an unconstrained schema.
		return &base.Schema{}
	}
}
