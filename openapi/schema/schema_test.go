package schema_test

import (
	"reflect"
	"testing"
	"time"

	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

func TestSchemaFromType_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		t        reflect.Type
		wantType string
	}{
		{"string", reflect.TypeFor[string](), "string"},
		{"bool", reflect.TypeFor[bool](), "boolean"},
		{"int", reflect.TypeFor[int](), "integer"},
		{"int8", reflect.TypeFor[int8](), "integer"},
		{"int16", reflect.TypeFor[int16](), "integer"},
		{"int32", reflect.TypeFor[int32](), "integer"},
		{"int64", reflect.TypeFor[int64](), "integer"},
		{"uint", reflect.TypeFor[uint](), "integer"},
		{"uint8", reflect.TypeFor[uint8](), "integer"},
		{"uint16", reflect.TypeFor[uint16](), "integer"},
		{"uint32", reflect.TypeFor[uint32](), "integer"},
		{"uint64", reflect.TypeFor[uint64](), "integer"},
		{"float32", reflect.TypeFor[float32](), "number"},
		{"float64", reflect.TypeFor[float64](), "number"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := schema.SchemaFromType(tt.t)
			if len(s.Type) != 1 || s.Type[0] != tt.wantType {
				t.Errorf("expected type [%s], got %v", tt.wantType, s.Type)
			}
		})
	}
}

func TestSchemaFromType_Pointer(t *testing.T) {
	s := schema.SchemaFromType(reflect.TypeFor[*string]())
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
	if s.Nullable == nil || !*s.Nullable {
		t.Error("expected nullable=true for pointer type")
	}
}

func TestSchemaFromType_Slice(t *testing.T) {
	s := schema.SchemaFromType(reflect.TypeFor[[]string]())
	if len(s.Type) != 1 || s.Type[0] != "array" {
		t.Errorf("expected type [array], got %v", s.Type)
	}
	if s.Items == nil || s.Items.A == nil {
		t.Fatal("expected items to be set")
	}
	items := s.Items.A.Schema()
	if len(items.Type) != 1 || items.Type[0] != "string" {
		t.Errorf("expected items type [string], got %v", items.Type)
	}
}

func TestSchemaFromType_Array(t *testing.T) {
	s := schema.SchemaFromType(reflect.TypeFor[[3]int]())
	if len(s.Type) != 1 || s.Type[0] != "array" {
		t.Errorf("expected type [array], got %v", s.Type)
	}
	if s.Items == nil || s.Items.A == nil {
		t.Fatal("expected items to be set")
	}
	items := s.Items.A.Schema()
	if len(items.Type) != 1 || items.Type[0] != "integer" {
		t.Errorf("expected items type [integer], got %v", items.Type)
	}
}

func TestSchemaFromType_Map(t *testing.T) {
	s := schema.SchemaFromType(reflect.TypeFor[map[string]int]())
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected type [object], got %v", s.Type)
	}
	if s.AdditionalProperties == nil || s.AdditionalProperties.N != 0 || s.AdditionalProperties.A == nil {
		t.Fatal("expected additionalProperties to be set as schema proxy")
	}
	addl := s.AdditionalProperties.A.Schema()
	if len(addl.Type) != 1 || addl.Type[0] != "integer" {
		t.Errorf("expected additionalProperties type [integer], got %v", addl.Type)
	}
}

func TestSchemaFromType_TimeTime(t *testing.T) {
	s := schema.SchemaFromType(reflect.TypeFor[time.Time]())
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
	if s.Format != "date-time" {
		t.Errorf("expected format date-time, got %q", s.Format)
	}
}

func TestSchemaFromType_Struct(t *testing.T) {
	type User struct {
		ID       int
		Name     string
		Email    string  `json:"email"`
		Password string  `json:"-"`
		Bio      *string `json:"bio"`
		internal string  //nolint:unused
	}

	s := schema.SchemaFromType(reflect.TypeFor[User]())

	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected type [object], got %v", s.Type)
	}

	// Properties that should be present.
	for _, name := range []string{"ID", "Name", "email", "bio"} {
		if _, ok := s.Properties.Get(name); !ok {
			t.Errorf("expected property %q to exist", name)
		}
	}

	// json:"-" and unexported fields should be excluded.
	for _, name := range []string{"Password", "internal"} {
		if _, ok := s.Properties.Get(name); ok {
			t.Errorf("expected property %q to be excluded", name)
		}
	}

	// Non-pointer fields are required.
	required := make(map[string]bool, len(s.Required))
	for _, r := range s.Required {
		required[r] = true
	}
	for _, name := range []string{"ID", "Name", "email"} {
		if !required[name] {
			t.Errorf("expected field %q to be required", name)
		}
	}
	if required["bio"] {
		t.Error("expected 'bio' (pointer type) to not be required")
	}

	// Struct should disallow additional properties.
	if s.AdditionalProperties == nil || s.AdditionalProperties.N != 1 || s.AdditionalProperties.B {
		t.Error("expected additionalProperties=false for struct")
	}
}

func TestSchemaFromType_StructPropertyTypes(t *testing.T) {
	type Item struct {
		Count int
		Name  string
		Score float64
		Tags  []string
	}

	s := schema.SchemaFromType(reflect.TypeFor[Item]())

	scalar := []struct {
		field    string
		wantType string
	}{
		{"Count", "integer"},
		{"Name", "string"},
		{"Score", "number"},
	}
	for _, c := range scalar {
		prop, ok := s.Properties.Get(c.field)
		if !ok {
			t.Errorf("expected property %q to exist", c.field)
			continue
		}
		if pt := prop.Schema().Type; len(pt) != 1 || pt[0] != c.wantType {
			t.Errorf("field %q: expected type [%s], got %v", c.field, c.wantType, pt)
		}
	}

	tagsProp, ok := s.Properties.Get("Tags")
	if !ok {
		t.Fatal("expected 'Tags' property to exist")
	}
	if pt := tagsProp.Schema().Type; len(pt) != 1 || pt[0] != "array" {
		t.Errorf("expected Tags type [array], got %v", pt)
	}
}

func TestSchemaFromType_StructPointerField(t *testing.T) {
	type Profile struct {
		Name string
		Bio  *string
	}

	s := schema.SchemaFromType(reflect.TypeFor[Profile]())

	bioProp, ok := s.Properties.Get("Bio")
	if !ok {
		t.Fatal("expected 'Bio' property to exist")
	}
	bioSchema := bioProp.Schema()
	if len(bioSchema.Type) != 1 || bioSchema.Type[0] != "string" {
		t.Errorf("expected Bio type [string], got %v", bioSchema.Type)
	}
	if bioSchema.Nullable == nil || !*bioSchema.Nullable {
		t.Error("expected Bio nullable=true for pointer field")
	}
	for _, r := range s.Required {
		if r == "Bio" {
			t.Error("expected 'Bio' (pointer type) to not be in required list")
		}
	}
}

func TestSchemaFromType_Recursive(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}
	// Should not loop infinitely.
	s := schema.SchemaFromType(reflect.TypeFor[Node]())
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected type [object], got %v", s.Type)
	}
	if _, ok := s.Properties.Get("Value"); !ok {
		t.Error("expected 'Value' property to exist")
	}
}

func TestSchemaFromType_UnsupportedType(t *testing.T) {
	// Channels fall back to an unconstrained schema.
	s := schema.SchemaFromType(reflect.TypeFor[chan int]())
	if len(s.Type) != 0 {
		t.Errorf("expected unconstrained schema (no type), got %v", s.Type)
	}
}

// reflectSchemaType is a struct type that implements SchemaOption and requests
// registration under a named $ref.
type reflectSchemaType struct {
	ID   int
	Name string
}

func (reflectSchemaType) Schema() schema.Option {
	return schema.Ref("ReflectSchemaType")
}

func TestReflectType_PointerNullable(t *testing.T) {
	proxy := schema.ReflectType(reflect.TypeFor[*string](), nil)
	s := proxy.Schema()
	if s == nil {
		t.Fatal("expected non-nil schema")
	}
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
	if s.Nullable == nil || !*s.Nullable {
		t.Error("expected nullable=true for pointer type")
	}
}

func TestReflectType_NilRegistry(t *testing.T) {
	proxy := schema.ReflectType(reflect.TypeFor[string](), nil)
	s := proxy.Schema()
	if s == nil {
		t.Fatal("expected non-nil schema for inline proxy")
	}
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
}

func TestReflectType_RegistryWithoutRef(t *testing.T) {
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	reg := schema.NewRegistry(schemas)

	proxy := schema.ReflectType(reflect.TypeFor[int](), reg)
	s := proxy.Schema()
	if s == nil {
		t.Fatal("expected non-nil schema for inline proxy")
	}
	if len(s.Type) != 1 || s.Type[0] != "integer" {
		t.Errorf("expected type [integer], got %v", s.Type)
	}

	// No schema should be stored in the map for a primitive without Ref.
	if _, ok := schemas.Get("int"); ok {
		t.Error("expected no schemas stored for primitive type without Ref")
	}
}

func TestReflectType_RegistryWithRef(t *testing.T) {
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	reg := schema.NewRegistry(schemas)

	schema.ReflectType(reflect.TypeFor[reflectSchemaType](), reg)

	// Schema must be stored in components/schemas under the Ref name.
	stored, ok := schemas.Get("ReflectSchemaType")
	if !ok {
		t.Fatal("expected schema stored in registry under 'ReflectSchemaType'")
	}
	storedSchema := stored.Schema()
	if storedSchema == nil {
		t.Fatal("expected stored schema to be non-nil")
	}
	if len(storedSchema.Type) != 1 || storedSchema.Type[0] != "object" {
		t.Errorf("expected stored schema type [object], got %v", storedSchema.Type)
	}
}

func TestReflectType_RegistryWithRef_Caching(t *testing.T) {
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	reg := schema.NewRegistry(schemas)

	schema.ReflectType(reflect.TypeFor[reflectSchemaType](), reg)
	schema.ReflectType(reflect.TypeFor[reflectSchemaType](), reg)

	// Schema should be stored exactly once despite two calls.
	if _, ok := schemas.Get("ReflectSchemaType"); !ok {
		t.Fatal("expected schema stored in registry")
	}
	// A second entry should not appear.
	count := 0
	for pair := schemas.First(); pair != nil; pair = pair.Next() {
		if pair.Key() == "ReflectSchemaType" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("expected exactly 1 registry entry for ReflectSchemaType, got %d", count)
	}
}

func TestReflectType_DereferencesPointer(t *testing.T) {
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	reg := schema.NewRegistry(schemas)

	// *reflectSchemaType should dereference and trigger the SchemaOption.
	schema.ReflectType(reflect.TypeFor[*reflectSchemaType](), reg)

	if _, ok := schemas.Get("ReflectSchemaType"); !ok {
		t.Error("expected pointer type to be dereferenced and trigger SchemaOption Ref")
	}
}

func TestReflectType_SchemaOptionApplied(t *testing.T) {
	type describedType struct {
		Value string
	}
	// Use a local SchemaOption that sets Description.
	// Since we can't define a method on a local type, we test via a named type.
	// Instead, verify that Ref is applied through the global reflectSchemaType.
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	reg := schema.NewRegistry(schemas)

	schema.ReflectType(reflect.TypeFor[reflectSchemaType](), reg)

	stored, ok := schemas.Get("ReflectSchemaType")
	if !ok {
		t.Fatal("expected schema stored")
	}
	// The stored schema is for reflectSchemaType which has ID and Name fields.
	s := stored.Schema()
	if _, ok := s.Properties.Get("ID"); !ok {
		t.Error("expected 'ID' property in stored schema")
	}
	if _, ok := s.Properties.Get("Name"); !ok {
		t.Error("expected 'Name' property in stored schema")
	}
}
