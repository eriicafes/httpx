package schema

import (
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

// RegistryField holds metadata derived from schema options that lives alongside the JSON Schema.
type RegistryField struct {
	// Reference is the component name to register this type under in components/schemas.
	Reference string
}

// Registry tracks component schemas by Go type, writing them into a shared schemas map.
// Pass a non-nil Registry to ReflectType to enable auto-registration of types that declare
// a Reference via schema.Ref. Pass nil to always generate inline schemas.
type Registry struct {
	fields  map[reflect.Type]*RegistryField
	schemas *orderedmap.Map[string, *base.SchemaProxy]
}

// NewRegistry creates a Registry backed by the given schemas map.
// The caller is responsible for ensuring the map is also wired into the OpenAPI document's
// components/schemas.
func NewRegistry(schemas *orderedmap.Map[string, *base.SchemaProxy]) *Registry {
	return &Registry{
		fields:  make(map[reflect.Type]*RegistryField),
		schemas: schemas,
	}
}

// Get builds and optionally registers the schema for t. build is always called with a fresh
// RegistryField.
// If it sets f.Reference the schema is stored in components/schemas under
// that name and a $ref is returned.
// If f.Reference is empty the proxy is returned inline.
// If t was previously registered, the cached $ref is returned without calling build.
// If the registry is nil, build is called and its result is returned inline.
func (r *Registry) Get(t reflect.Type, build func(*RegistryField) *base.SchemaProxy) *base.SchemaProxy {
	f := &RegistryField{}
	if r == nil {
		return build(f)
	}
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if f, ok := r.fields[t]; ok {
		return base.CreateSchemaProxyRef(f.Reference)
	}

	proxy := build(f)
	if f.Reference == "" {
		return proxy
	}

	ref := "#/components/schemas/" + f.Reference
	r.fields[t] = &RegistryField{Reference: ref}
	r.schemas.Set(f.Reference, proxy)
	return base.CreateSchemaProxyRef(ref)
}
