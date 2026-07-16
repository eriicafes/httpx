package schema

import (
	"reflect"
	"strings"

	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// Schema may be implemented to set default schema options on a type.
//
//	func (T) Schema() schema.Option
type Schema interface {
	Schema() Option
}

// New reflects T into an OpenAPI schema proxy.
// When T declares schema options with Reference, the returned proxy is a $ref.
func New[T any](store *store.Store) *base.SchemaProxy {
	return getSchemaForType(reflect.TypeFor[T](), store, make(map[reflect.Type]bool))
}

// TypeGetter lazily resolves a schema proxy using the provided component store.
type TypeGetter func(store *store.Store) *base.SchemaProxy

// Type returns a TypeGetter for T for use with composition helpers such as OneOf and AllOf.
func Type[T any]() TypeGetter {
	return func(store *store.Store) *base.SchemaProxy {
		return getSchemaForType(reflect.TypeFor[T](), store, make(map[reflect.Type]bool))
	}
}

// TypeOf reflects the dynamic type of i into an OpenAPI schema proxy.
func TypeOf(i any, store *store.Store) *base.SchemaProxy {
	return getSchemaForType(reflect.TypeOf(i), store, make(map[reflect.Type]bool))
}

// getSchemaForType returns the JSON Schema for t.
// For pointer types it sets nullable on the element's schema.
func getSchemaForType(t reflect.Type, store *store.Store, visited map[reflect.Type]bool) *base.SchemaProxy {
	var isPointer bool
	for t.Kind() == reflect.Pointer {
		t, isPointer = t.Elem(), true
	}

	// Get schema proxy ref for type if stored
	if store != nil {
		if schema, ok := store.GetSchema(t); ok {
			return schema
		}
	}

	var schema *base.Schema
	var unsupported bool

	switch t.Kind() {
	case reflect.String:
		schema = &base.Schema{Type: []string{"string"}}
	case reflect.Bool:
		schema = &base.Schema{Type: []string{"boolean"}}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		schema = &base.Schema{Type: []string{"integer"}}
	case reflect.Float32, reflect.Float64:
		schema = &base.Schema{Type: []string{"number"}}
	case reflect.Slice:
		schema = &base.Schema{
			Type: []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: getSchemaForType(t.Elem(), store, visited),
			},
		}
	case reflect.Array:
		n := int64(t.Len())
		schema = &base.Schema{
			Type: []string{"array"},
			Items: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: getSchemaForType(t.Elem(), store, visited),
			},
			MinItems: &n,
			MaxItems: &n,
		}
	case reflect.Map:
		schema = &base.Schema{
			Type: []string{"object"},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 0,
				A: getSchemaForType(t.Elem(), store, visited),
			},
		}
	case reflect.Struct:
		// union types expand to oneOf / tagged oneOf with discriminator
		if info := unionFromType(t, store, visited); info != nil {
			schema = &base.Schema{AnyOf: info.schemas}
			if info.discriminator != nil {
				schema.AnyOf = nil
				schema.OneOf = info.schemas
				schema.Discriminator = info.discriminator
			}
			break
		}

		// special case time.Time
		if t.PkgPath() == "time" && t.Name() == "Time" {
			schema = &base.Schema{
				Type:   []string{"string"},
				Format: "date-time",
			}
			break
		}

		// Guard against self-referential types
		if visited[t] {
			schema = &base.Schema{}
			break
		}
		visited[t] = true

		props := orderedmap.New[string, *base.SchemaProxy]()
		required := collectStructFields(t, store, visited, props, nil)

		schema = &base.Schema{
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
		schema, unsupported = &base.Schema{}, true
	}

	// Apply default nullable for pointer types
	if isPointer {
		schema.Nullable = &isPointer
	}

	if unsupported {
		return base.CreateSchemaProxy(schema)
	}

	// Apply schema options via interface
	var state State
	if sc, ok := reflect.New(t).Interface().(Schema); ok {
		sc.Schema()(schema, &Store{store, &state})
	}

	proxy := base.CreateSchemaProxy(schema)
	// If a reference is set, store the full schema in components
	// and return a schema $ref object.
	if store != nil && state.Reference != "" {
		return store.SetSchema(t, state.Reference, proxy)
	}
	return proxy
}

// collectStructFields appends t's fields as properties into props and returns
// required with the names of required fields appended.
// Untagged anonymous struct fields (or pointers to structs) are flattened into
// the parent object, matching encoding/json promotion semantics.
func collectStructFields(t reflect.Type, store *store.Store, visited map[reflect.Type]bool, props *orderedmap.Map[string, *base.SchemaProxy], required []string) []string {
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		tag, hasTag := f.Tag.Lookup("json")
		if f.Anonymous && !hasTag {
			ft := f.Type
			for ft.Kind() == reflect.Pointer {
				ft = ft.Elem()
			}
			isTime := ft.PkgPath() == "time" && ft.Name() == "Time"
			if ft.Kind() == reflect.Struct && !isTime {
				required = collectStructFields(ft, store, visited, props, required)
				continue
			}
		}

		name, omittable := f.Name, false
		if hasTag {
			tagName, restTag, _ := strings.Cut(tag, ",")
			if tagName == "-" {
				continue
			}
			if tagName != "" {
				name = tagName
			}
			if strings.Contains(restTag, "omitempty") || strings.Contains(restTag, "omitzero") {
				omittable = true
			}
		}
		fieldSchema := getSchemaForType(f.Type, store, visited)
		if fieldSchema == nil {
			continue
		}
		props.Set(name, fieldSchema)
		if f.Type.Kind() != reflect.Pointer && !omittable {
			required = append(required, name)
		}
	}
	return required
}

const unionPkg = "github.com/eriicafes/union"

type unionInfo struct {
	schemas       []*base.SchemaProxy
	discriminator *base.Discriminator
}

// unionFromType checks if t is a union.Union or union.TaggedUnion and returns
// the expanded case schemas with an optional discriminator. Returns nil if t is
// not a union type.
func unionFromType(t reflect.Type, store *store.Store, visited map[reflect.Type]bool) *unionInfo {
	if t.PkgPath() != unionPkg {
		return nil
	}

	name := t.Name()
	isTagged := strings.HasPrefix(name, "TaggedUnion")
	isUntagged := strings.HasPrefix(name, "Union")
	if !isTagged && !isUntagged {
		return nil
	}

	// Both Union[Spec] and TaggedUnion[Spec] are struct{ Value Spec }.
	if t.NumField() != 1 {
		return nil
	}
	specType := t.Field(0).Type
	if specType.Kind() != reflect.Struct {
		return nil
	}

	if isTagged {
		// Resolve field names via interface
		variantKey, valueKey := "type", "value"
		if spec, ok := reflect.New(specType).Interface().(interface {
			TaggedFieldNames() (string, string)
		}); ok {
			variantKey, valueKey = spec.TaggedFieldNames()
		}
		return taggedUnionFromSpec(specType, variantKey, valueKey, store, visited)
	}
	return unionFromSpec(specType, store, visited)
}

// unionFromSpec builds case schemas for a Union.
func unionFromSpec(specType reflect.Type, store *store.Store, visited map[reflect.Type]bool) *unionInfo {
	info := &unionInfo{}
	for i := range specType.NumField() {
		f := specType.Field(i)
		info.schemas = append(info.schemas, getSchemaForType(f.Type, store, visited))
	}
	return info
}

// taggedUnionFromSpec builds case schemas for a TaggedUnion.
// Each case is an object with a const discriminator property (variantKey) and a
// value property (valueKey) holding the field's schema.
func taggedUnionFromSpec(specType reflect.Type, variantKey, valueKey string, store *store.Store, visited map[reflect.Type]bool) *unionInfo {
	info := &unionInfo{
		discriminator: &base.Discriminator{PropertyName: variantKey},
	}

	for i := range specType.NumField() {
		f := specType.Field(i)
		variantName := f.Tag.Get("variant")
		if variantName == "" {
			variantName = f.Name
		}

		props := orderedmap.New[string, *base.SchemaProxy]()
		props.Set(variantKey, base.CreateSchemaProxy(&base.Schema{
			Const: ToYAMLNode(variantName),
		}))
		props.Set(valueKey, getSchemaForType(f.Type, store, visited))

		caseSchema := &base.Schema{
			Type:       []string{"object"},
			Properties: props,
			Required:   []string{variantKey, valueKey},
			AdditionalProperties: &base.DynamicValue[*base.SchemaProxy, bool]{
				N: 1, B: false,
			},
		}
		info.schemas = append(info.schemas, base.CreateSchemaProxy(caseSchema))
	}

	return info
}

// ToYAMLNode converts v into a YAML node suitable for libopenapi example-like fields.
func ToYAMLNode(v any) *yaml.Node {
	node := &yaml.Node{}
	b, _ := yaml.Marshal(v)
	_ = yaml.Unmarshal(b, node)
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		return node.Content[0]
	}
	return node
}
