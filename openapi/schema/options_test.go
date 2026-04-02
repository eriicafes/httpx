package schema_test

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

func applySchemaOptions(s *base.Schema, opts ...schema.Option) {
	st := &schema.Store{
		Store: store.New(&v3.Components{}),
	}
	for _, opt := range opts {
		opt(s, st)
	}
}

func TestOptions_GeneralScalarAndXML(t *testing.T) {
	s := &base.Schema{}
	applySchemaOptions(s,
		schema.Title("User"),
		schema.Description("A user object"),
		schema.Format("uuid"),
		schema.ReadOnly(),
		schema.WriteOnly(),
		schema.Deprecated(),
		schema.Nullable(),
		schema.ExternalDocs("https://example.com/user", "details"),
		schema.Default("guest"),
		schema.Example("ada"),
		schema.Examples("ada", "grace"),
		schema.Const("fixed"),
		schema.Enum("admin", "user"),
		schema.MinLength(3),
		schema.MaxLength(20),
		schema.Pattern("^[a-z]+$"),
		schema.ContentEncoding("base64"),
		schema.ContentMediaType("image/png"),
		schema.MultipleOf(2),
		schema.Minimum(1),
		schema.Maximum(10),
		schema.ExclusiveMinimum(2),
		schema.ExclusiveMaximum(9),
		schema.XMLName("user"),
		schema.XMLNamespace("urn:users"),
		schema.XMLPrefix("u"),
		schema.XMLAttribute(),
		schema.XMLWrapped(),
	)

	if s.Title != "User" || s.Description != "A user object" || s.Format != "uuid" {
		t.Fatalf("general schema fields not applied: %#v", s)
	}
	if s.ReadOnly == nil || !*s.ReadOnly || s.WriteOnly == nil || !*s.WriteOnly || s.Deprecated == nil || !*s.Deprecated || s.Nullable == nil || !*s.Nullable {
		t.Fatalf("schema flags not applied: %#v", s)
	}
	if s.ExternalDocs == nil || s.ExternalDocs.URL != "https://example.com/user" || s.ExternalDocs.Description != "details" {
		t.Fatalf("external docs not applied: %#v", s.ExternalDocs)
	}
	if s.Default == nil || s.Default.Value != "guest" || s.Example == nil || s.Example.Value != "ada" {
		t.Fatalf("default/example not applied: default=%#v example=%#v", s.Default, s.Example)
	}
	if len(s.Examples) != 2 || s.Examples[0].Value != "ada" || s.Examples[1].Value != "grace" {
		t.Fatalf("examples not applied: %#v", s.Examples)
	}
	if s.Const == nil || s.Const.Value != "fixed" || len(s.Enum) != 2 || s.Enum[0].Value != "admin" || s.Enum[1].Value != "user" {
		t.Fatalf("const/enum not applied: const=%#v enum=%#v", s.Const, s.Enum)
	}
	if s.MinLength == nil || *s.MinLength != 3 || s.MaxLength == nil || *s.MaxLength != 20 || s.Pattern != "^[a-z]+$" {
		t.Fatalf("string constraints not applied: %#v", s)
	}
	if s.ContentEncoding != "base64" || s.ContentMediaType != "image/png" {
		t.Fatalf("content metadata not applied: %#v", s)
	}
	if s.MultipleOf == nil || *s.MultipleOf != 2 || s.Minimum == nil || *s.Minimum != 1 || s.Maximum == nil || *s.Maximum != 10 {
		t.Fatalf("numeric constraints not applied: %#v", s)
	}
	if s.ExclusiveMinimum == nil || s.ExclusiveMinimum.N != 1 || s.ExclusiveMinimum.B != 2 {
		t.Fatalf("exclusive minimum not applied: %#v", s.ExclusiveMinimum)
	}
	if s.ExclusiveMaximum == nil || s.ExclusiveMaximum.N != 1 || s.ExclusiveMaximum.B != 9 {
		t.Fatalf("exclusive maximum not applied: %#v", s.ExclusiveMaximum)
	}
	if s.XML == nil || s.XML.Name != "user" || s.XML.Namespace != "urn:users" || s.XML.Prefix != "u" || !s.XML.Attribute || !s.XML.Wrapped {
		t.Fatalf("xml options not applied: %#v", s.XML)
	}
}

func TestOptions_ArrayObjectAndConditionals(t *testing.T) {
	type Profile struct {
		Name string
		Tags []string
	}

	proxy := schema.New[Profile](store.New(&v3.Components{}))
	s := proxy.Schema()
	applySchemaOptions(s,
		schema.MinItems(1),
		schema.MaxItems(5),
		schema.UniqueItems(),
		schema.MinContains(1),
		schema.MaxContains(2),
		schema.PrefixItems(schema.Type[string](), schema.Type[int]()),
		schema.Contains[string](),
		schema.UnevaluatedItems[int](),
		schema.MinProperties(1),
		schema.MaxProperties(4),
		schema.AdditionalProperties[string](),
		schema.Field("Name", schema.MinLength(2), schema.Description("display name")),
		schema.PatternProperties("^x-", schema.MinLength(1)),
		schema.Required("Name"),
		schema.DependentRequired("Name", "Tags"),
		schema.DependentSchemas[string]("Name"),
		schema.PropertyNames[string](),
		schema.UnevaluatedProperties[int](),
		schema.Not[bool](),
		schema.If[string](),
		schema.Then[int](),
		schema.Else[bool](),
	)

	if s.MinItems == nil || *s.MinItems != 1 || s.MaxItems == nil || *s.MaxItems != 5 || s.UniqueItems == nil || !*s.UniqueItems {
		t.Fatalf("array constraints not applied: %#v", s)
	}
	if s.MinContains == nil || *s.MinContains != 1 || s.MaxContains == nil || *s.MaxContains != 2 {
		t.Fatalf("contains constraints not applied: %#v", s)
	}
	if len(s.PrefixItems) != 2 {
		t.Fatalf("prefix items not applied: %#v", s.PrefixItems)
	}
	if s.Contains == nil || s.Contains.Schema() == nil || len(s.Contains.Schema().Type) != 1 || s.Contains.Schema().Type[0] != "string" {
		t.Fatalf("contains schema not applied: %#v", s.Contains)
	}
	if s.UnevaluatedItems == nil || s.UnevaluatedItems.Schema() == nil || s.UnevaluatedItems.Schema().Type[0] != "integer" {
		t.Fatalf("unevaluated items not applied: %#v", s.UnevaluatedItems)
	}
	if s.MinProperties == nil || *s.MinProperties != 1 || s.MaxProperties == nil || *s.MaxProperties != 4 {
		t.Fatalf("object property count constraints not applied: %#v", s)
	}
	if s.AdditionalProperties == nil || s.AdditionalProperties.N != 0 || s.AdditionalProperties.A == nil || s.AdditionalProperties.A.Schema().Type[0] != "string" {
		t.Fatalf("additional properties not applied: %#v", s.AdditionalProperties)
	}
	nameProp, ok := s.Properties.Get("Name")
	if !ok {
		t.Fatal("Name property missing")
	}
	if nameProp.Schema().MinLength == nil || *nameProp.Schema().MinLength != 2 || nameProp.Schema().Description != "display name" {
		t.Fatalf("field options not applied: %#v", nameProp.Schema())
	}
	pp, ok := s.PatternProperties.Get("^x-")
	if !ok || pp.Schema().MinLength == nil || *pp.Schema().MinLength != 1 {
		t.Fatalf("pattern properties not applied: %#v ok=%v", pp, ok)
	}
	if len(s.Required) != 1 || s.Required[0] != "Name" {
		t.Fatalf("required override not applied: %#v", s.Required)
	}
	dr, ok := s.DependentRequired.Get("Name")
	if !ok || len(dr) != 1 || dr[0] != "Tags" {
		t.Fatalf("dependent required not applied: %#v ok=%v", dr, ok)
	}
	ds, ok := s.DependentSchemas.Get("Name")
	if !ok || ds.Schema() == nil || ds.Schema().Type[0] != "string" {
		t.Fatalf("dependent schema not applied: %#v ok=%v", ds, ok)
	}
	if s.PropertyNames == nil || s.PropertyNames.Schema() == nil || s.PropertyNames.Schema().Type[0] != "string" {
		t.Fatalf("property names not applied: %#v", s.PropertyNames)
	}
	if s.UnevaluatedProperties == nil || s.UnevaluatedProperties.N != 0 || s.UnevaluatedProperties.A == nil || s.UnevaluatedProperties.A.Schema().Type[0] != "integer" {
		t.Fatalf("unevaluated properties not applied: %#v", s.UnevaluatedProperties)
	}
	if s.Not == nil || s.Not.Schema() == nil || s.Not.Schema().Type[0] != "boolean" {
		t.Fatalf("not schema not applied: %#v", s.Not)
	}
	if s.If == nil || s.If.Schema() == nil || s.If.Schema().Type[0] != "string" {
		t.Fatalf("if schema not applied: %#v", s.If)
	}
	if s.Then == nil || s.Then.Schema() == nil || s.Then.Schema().Type[0] != "integer" {
		t.Fatalf("then schema not applied: %#v", s.Then)
	}
	if s.Else == nil || s.Else.Schema() == nil || s.Else.Schema().Type[0] != "boolean" {
		t.Fatalf("else schema not applied: %#v", s.Else)
	}
}

func TestOptions_CompositionHelpers(t *testing.T) {
	s := &base.Schema{}
	applySchemaOptions(s,
		schema.OneOf(schema.Type[bool]()),
		schema.AnyOf(schema.Type[string](), schema.Type[int]()),
		schema.AllOf(schema.Type[string](), schema.Type[int]()),
	)
	if len(s.OneOf) != 1 || len(s.AnyOf) != 2 || len(s.AllOf) != 2 {
		t.Fatalf("composition helpers not applied: oneOf=%d anyOf=%d allOf=%d", len(s.OneOf), len(s.AnyOf), len(s.AllOf))
	}
}

func TestToYAMLNode(t *testing.T) {
	scalar := schema.ToYAMLNode("ada")
	if scalar.Kind != yaml.ScalarNode || scalar.Value != "ada" {
		t.Fatalf("scalar node mismatch: %#v", scalar)
	}

	mapping := schema.ToYAMLNode(map[string]any{"name": "ada", "age": 10})
	if mapping.Kind != yaml.MappingNode || len(mapping.Content) != 4 {
		t.Fatalf("mapping node mismatch: kind=%d content=%d", mapping.Kind, len(mapping.Content))
	}

	seq := schema.ToYAMLNode([]string{"a", "b"})
	if seq.Kind != yaml.SequenceNode || len(seq.Content) != 2 {
		t.Fatalf("sequence node mismatch: kind=%d content=%d", seq.Kind, len(seq.Content))
	}
}

func TestField_NoOpWhenMissing(t *testing.T) {
	s := &base.Schema{Properties: orderedmap.New[string, *base.SchemaProxy]()}
	applySchemaOptions(s, schema.Field("Missing", schema.MinLength(2)))
	if s.Properties.Len() != 0 {
		t.Fatalf("missing field should be a no-op: %#v", s.Properties)
	}
}
