package schema

import (
	"testing"
	"time"

	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNew_Primitives(t *testing.T) {
	tests := []struct {
		name     string
		v        any
		wantType string
	}{
		{"string", "", "string"},
		{"bool", false, "boolean"},
		{"int", int(0), "integer"},
		{"int8", int8(0), "integer"},
		{"int16", int16(0), "integer"},
		{"int32", int32(0), "integer"},
		{"int64", int64(0), "integer"},
		{"uint", uint(0), "integer"},
		{"uint8", uint8(0), "integer"},
		{"uint16", uint16(0), "integer"},
		{"uint32", uint32(0), "integer"},
		{"uint64", uint64(0), "integer"},
		{"float32", float32(0), "number"},
		{"float64", float64(0), "number"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxy := TypeOf(tt.v, nil)
			s := proxy.Schema()
			if len(s.Type) != 1 || s.Type[0] != tt.wantType {
				t.Errorf("expected type [%s], got %v", tt.wantType, s.Type)
			}
		})
	}
}

func TestNew_Recursive(t *testing.T) {
	type Node struct {
		Value int
		Next  *Node
	}
	// Should not loop infinitely.
	proxy := New[Node](nil)
	s := proxy.Schema()
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected type [object], got %v", s.Type)
	}
	if _, ok := s.Properties.Get("Value"); !ok {
		t.Error("expected 'Value' property to exist")
	}
}

func TestNew_UnsupportedType(t *testing.T) {
	// Channels fall back to an unconstrained schema.
	proxy := New[chan int](nil)
	s := proxy.Schema()
	if len(s.Type) != 0 {
		t.Errorf("expected unconstrained schema (no type), got %v", s.Type)
	}
}

func TestNew_PointerNullable(t *testing.T) {
	proxy := New[*string](nil)
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

func TestNew_NilRegistry(t *testing.T) {
	proxy := New[string](nil)
	s := proxy.Schema()
	if s == nil {
		t.Fatal("expected non-nil schema for inline proxy")
	}
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
}

func TestNew_Pointer(t *testing.T) {
	proxy := New[*string](nil)
	s := proxy.Schema()
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
	if s.Nullable == nil || !*s.Nullable {
		t.Error("expected nullable=true for pointer type")
	}
}

func TestNew_Slice(t *testing.T) {
	proxy := New[[]string](nil)
	s := proxy.Schema()
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

func TestNew_Array(t *testing.T) {
	proxy := New[[3]int](nil)
	s := proxy.Schema()
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

func TestNew_Map(t *testing.T) {
	proxy := New[map[string]int](nil)
	s := proxy.Schema()
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

func TestNew_Time(t *testing.T) {
	proxy := New[time.Time](nil)
	s := proxy.Schema()
	if len(s.Type) != 1 || s.Type[0] != "string" {
		t.Errorf("expected type [string], got %v", s.Type)
	}
	if s.Format != "date-time" {
		t.Errorf("expected format date-time, got %q", s.Format)
	}
}

func TestNew_Struct(t *testing.T) {
	type User struct {
		ID       int
		Name     string
		Email    string  `json:"email"`
		Password string  `json:"-"`
		Bio      *string `json:"bio"`
		Age      int     `json:"age,omitzero"`
		Role     string  `json:"role,omitempty"`
		internal string  //nolint:unused
	}

	proxy := New[User](nil)
	s := proxy.Schema()

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
	if required["age"] {
		t.Error("expected 'age' (omitzero field) to not be required")
	}
	if required["role"] {
		t.Error("expected 'role' (omitempty field) to not be required")
	}

	// Struct should disallow additional properties.
	if s.AdditionalProperties == nil || s.AdditionalProperties.N != 1 || s.AdditionalProperties.B {
		t.Error("expected additionalProperties=false for struct")
	}
}

func TestNew_StructPropertyTypes(t *testing.T) {
	type Item struct {
		Count int
		Name  string
		Score float64
		Tags  []string
	}

	proxy := New[Item](nil)
	s := proxy.Schema()

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

func TestNew_StructPointerField(t *testing.T) {
	type Profile struct {
		Name string
		Bio  *string
	}

	proxy := New[Profile](nil)
	s := proxy.Schema()

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

// schemaType is a struct type that implements Schema and requests
// registration under a named $ref.
type schemaType struct {
	ID   int
	Name string
}

func (schemaType) Schema() Option {
	return Reference("SchemaType")
}

type nonSchemaType struct {
	ID   int
	Name string
}

func TestNew_WithoutReference(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	// Types without a Reference must be returned as an inline schema, not a $ref.
	proxy := New[nonSchemaType](store)
	s := proxy.Schema()
	if s == nil {
		t.Fatal("expected inline schema (non-nil Schema()), got $ref proxy")
	}
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected type [object], got %v", s.Type)
	}

	// No schema should be stored in the map for a type without Reference.
	if components.Schemas != nil && components.Schemas.Len() > 0 {
		t.Error("expected no schemas stored for type without Reference")
	}
}

func TestNew_ReferenceIsNeverInlined(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	// First call: schema is built and stored; a $ref proxy must be returned, not the inline schema.
	proxy := New[schemaType](store)
	if proxy.Schema() != nil {
		t.Error("first call: expected $ref proxy (Schema() must be nil without document context), got inline schema")
	}

	// Second call: type is already in the store; must still return a $ref proxy.
	proxy2 := New[schemaType](store)
	if proxy2.Schema() != nil {
		t.Error("second call: expected $ref proxy, got inline schema")
	}

	// The full schema must be present in components/schemas.
	name := "SchemaType"
	stored, ok := components.Schemas.Get(name)
	if !ok {
		t.Fatalf("expected schema stored in components under '%s'", name)
	}
	s := stored.Schema()
	if s == nil {
		t.Fatal("expected stored schema to be non-nil")
	}
	if len(s.Type) != 1 || s.Type[0] != "object" {
		t.Errorf("expected stored schema type [object], got %v", s.Type)
	}
}

func TestNew_StoresReference(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	New[schemaType](store)

	// Schema must be stored in components/schemas.
	name := "SchemaType"
	stored, ok := components.Schemas.Get(name)
	if !ok {
		t.Fatalf("expected schema stored in registry under '%s'", name)
	}
	storedSchema := stored.Schema()
	if storedSchema == nil {
		t.Fatal("expected stored schema to be non-nil")
	}
	if len(storedSchema.Type) != 1 || storedSchema.Type[0] != "object" {
		t.Errorf("expected stored schema type [object], got %v", storedSchema.Type)
	}
}

func TestNew_ReturnsReference(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	New[schemaType](store).Schema()
	New[schemaType](store).Schema()

	// Schema should be stored exactly once despite two calls.
	name := "SchemaType"
	if _, ok := components.Schemas.Get(name); !ok {
		t.Fatalf("expected schema stored in registry under '%s'", name)
	}
	// Only one entry should not stored.
	if components.Schemas.Len() != 1 {
		t.Errorf("expected exactly 1 registry entry, got %d", components.Schemas.Len())
	}
}

func TestNew_DereferencesPointer(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	// *schemaType should dereference and trigger the Schema.
	New[*schemaType](store)

	name := "SchemaType"
	if _, ok := components.Schemas.Get(name); !ok {
		t.Error("expected pointer type to be dereferenced and trigger Schema Reference")
	}
}

func TestNew_DereferencesDoublePointer(t *testing.T) {
	components := &v3.Components{}
	store := store.New(components)

	// *schemaType should dereference and trigger the Schema.
	New[**schemaType](store)

	name := "SchemaType"
	if _, ok := components.Schemas.Get(name); !ok {
		t.Error("expected pointer type to be dereferenced and trigger Schema Reference")
	}
}
