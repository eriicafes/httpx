package schema

import (
	"reflect"

	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

type Option func(*base.Schema, *Store)

type Store struct {
	*store.Store
	state *State
}

type State struct {
	Reference string
}

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(s *base.Schema, store *Store) {
		for _, opt := range opts {
			opt(s, store)
		}
	}
}

// Reference sets a $ref to a named schema component in components/schemas.
func Reference(name string) Option {
	return func(_ *base.Schema, store *Store) {
		store.state.Reference = name
	}
}

// general

// Title sets the schema title.
func Title(s string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Title = s
	}
}

// Description sets the schema description.
func Description(s string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Description = s
	}
}

// Format sets the format hint for the schema (e.g. "date-time", "email", "uuid").
// Prefer the named helpers (Email, UUID, DateTime) for common formats.
func Format(s string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Format = s
	}
}

// ReadOnly marks the schema as read-only — the value must not be sent in requests. Default: false.
func ReadOnly() Option {
	v := true
	return func(schema *base.Schema, _ *Store) {
		schema.ReadOnly = &v
	}
}

// WriteOnly marks the schema as write-only — the value must not be included in responses. Default: false.
func WriteOnly() Option {
	v := true
	return func(schema *base.Schema, _ *Store) {
		schema.WriteOnly = &v
	}
}

// Deprecated marks the schema as deprecated. Default: false.
func Deprecated() Option {
	v := true
	return func(schema *base.Schema, _ *Store) {
		schema.Deprecated = &v
	}
}

// Nullable marks the schema as nullable (OpenAPI 3.0 compatibility). Default: false.
// In OpenAPI 3.1 prefer using a pointer type, which sets nullable automatically.
func Nullable() Option {
	v := true
	return func(schema *base.Schema, _ *Store) {
		schema.Nullable = &v
	}
}

// ExternalDocs attaches an external documentation link to the schema.
func ExternalDocs(url, description string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Default sets the default value used when the field is absent.
func Default(v any) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Default = ToYAMLNode(v)
	}
}

// Example sets an example value for the schema.
func Example(v any) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Example = ToYAMLNode(v)
	}
}

// Examples sets the array of example values for the schema (OpenAPI 3.1).
func Examples(values ...any) Option {
	return func(s *base.Schema, _ *Store) {
		nodes := make([]*yaml.Node, len(values))
		for i, v := range values {
			nodes[i] = ToYAMLNode(v)
		}
		s.Examples = nodes
	}
}

// Const restricts the schema to a single constant value.
func Const(v any) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.Const = ToYAMLNode(v)
	}
}

// Enum restricts the schema to a fixed set of allowed values.
func Enum(values ...any) Option {
	return func(schema *base.Schema, _ *Store) {
		nodes := make([]*yaml.Node, len(values))
		for i, v := range values {
			nodes[i] = ToYAMLNode(v)
		}
		schema.Enum = nodes
	}
}

// string

// MinLength sets the minimum number of characters for a string value.
func MinLength(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MinLength = &n
	}
}

// MaxLength sets the maximum number of characters for a string value.
func MaxLength(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MaxLength = &n
	}
}

// Pattern restricts a string value to match the given regular expression.
func Pattern(p string) Option {
	return func(s *base.Schema, _ *Store) {
		s.Pattern = p
	}
}

// Email sets the format to "email".
func Email() Option {
	return func(s *base.Schema, _ *Store) {
		s.Format = "email"
	}
}

// UUID sets the format to "uuid".
func UUID() Option {
	return func(s *base.Schema, _ *Store) {
		s.Format = "uuid"
	}
}

// DateTime sets the format to "date-time".
func DateTime() Option {
	return func(s *base.Schema, _ *Store) {
		s.Format = "date-time"
	}
}

// ContentEncoding sets the encoding for string content (e.g. "base64", "quoted-printable").
func ContentEncoding(s string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.ContentEncoding = s
	}
}

// ContentMediaType sets the MIME type of string content (e.g. "image/png", "application/pdf").
func ContentMediaType(s string) Option {
	return func(schema *base.Schema, _ *Store) {
		schema.ContentMediaType = s
	}
}

// numeric

// MultipleOf requires the numeric value to be a multiple of the given number.
func MultipleOf(v float64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MultipleOf = &v
	}
}

// Minimum sets the inclusive lower bound for a numeric value.
func Minimum(v float64) Option {
	return func(s *base.Schema, _ *Store) {
		s.Minimum = &v
	}
}

// Maximum sets the inclusive upper bound for a numeric value.
func Maximum(v float64) Option {
	return func(s *base.Schema, _ *Store) {
		s.Maximum = &v
	}
}

// ExclusiveMinimum sets the exclusive lower bound for a numeric value (OpenAPI 3.1).
func ExclusiveMinimum(v float64) Option {
	return func(s *base.Schema, _ *Store) {
		s.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	}
}

// ExclusiveMaximum sets the exclusive upper bound for a numeric value (OpenAPI 3.1).
func ExclusiveMaximum(v float64) Option {
	return func(s *base.Schema, _ *Store) {
		s.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	}
}

// array

// MinItems sets the minimum number of elements in an array.
func MinItems(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MinItems = &n
	}
}

// MaxItems sets the maximum number of elements in an array.
func MaxItems(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MaxItems = &n
	}
}

// UniqueItems requires all elements in an array to be distinct. Default: false.
func UniqueItems() Option {
	v := true
	return func(s *base.Schema, _ *Store) {
		s.UniqueItems = &v
	}
}

// MinContains sets the minimum number of items that must validate against the contains schema.
func MinContains(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MinContains = &n
	}
}

// MaxContains sets the maximum number of items that may validate against the contains schema.
func MaxContains(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MaxContains = &n
	}
}

// PrefixItems sets the tuple validation schemas for each position in an array (OpenAPI 3.1).
// Pass zero values of each positional type (e.g. PrefixItems(int(0), "")).
func PrefixItems(types ...TypeGetter) Option {
	return func(s *base.Schema, store *Store) {
		schemas := make([]*base.SchemaProxy, len(types))
		for i, t := range types {
			schemas[i] = t(store.Store)
		}
		s.PrefixItems = schemas
	}
}

// Contains sets the schema that at least one item in the array must validate against (OpenAPI 3.1).
func Contains[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.Contains = New[T](store.Store)
	}
}

// UnevaluatedItems sets the schema for array items not covered by items or prefixItems (OpenAPI 3.1).
func UnevaluatedItems[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.UnevaluatedItems = New[T](store.Store)
	}
}

// object

// MinProperties sets the minimum number of properties an object must have.
func MinProperties(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MinProperties = &n
	}
}

// MaxProperties sets the maximum number of properties an object may have.
func MaxProperties(n int64) Option {
	return func(s *base.Schema, _ *Store) {
		s.MaxProperties = &n
	}
}

// AdditionalProperties overrides the default closed-object behavior for structs.
// Use AdditionalProperties[any]() to allow any extra fields, or a concrete type
// to constrain their schema.
func AdditionalProperties[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: New[T](store.Store),
		}
	}
}

// Field applies options to a named property of a struct schema.
// If the property does not exist the option is a no-op.
func Field(name string, opts ...Option) Option {
	return func(s *base.Schema, store *Store) {
		prop, ok := s.Properties.Get(name)
		if !ok {
			return
		}
		for _, opt := range opts {
			opt(prop.Schema(), &Store{Store: store.Store})
		}
	}
}

// PatternProperties adds schema constraints for properties whose names match the given regex.
func PatternProperties(pattern string, opts ...Option) Option {
	return func(s *base.Schema, store *Store) {
		if s.PatternProperties == nil {
			s.PatternProperties = orderedmap.New[string, *base.SchemaProxy]()
		}
		inner := &base.Schema{}
		for _, opt := range opts {
			opt(inner, &Store{Store: store.Store})
		}
		s.PatternProperties.Set(pattern, base.CreateSchemaProxy(inner))
	}
}

// Required overrides the required property list for an object schema.
// By default, non-pointer struct fields are required and pointer fields are optional.
func Required(fields ...string) Option {
	return func(s *base.Schema, _ *Store) {
		s.Required = fields
	}
}

// DependentRequired specifies properties that are required when the named property is present (OpenAPI 3.1).
func DependentRequired(property string, required ...string) Option {
	return func(s *base.Schema, _ *Store) {
		if s.DependentRequired == nil {
			s.DependentRequired = orderedmap.New[string, []string]()
		}
		s.DependentRequired.Set(property, required)
	}
}

// DependentSchemas sets the schema that must validate when the named property is present (OpenAPI 3.1).
func DependentSchemas[T any](property string) Option {
	return func(s *base.Schema, store *Store) {
		if s.DependentSchemas == nil {
			s.DependentSchemas = orderedmap.New[string, *base.SchemaProxy]()
		}
		s.DependentSchemas.Set(property, New[T](store.Store))
	}
}

// PropertyNames sets constraints on the names of object properties (OpenAPI 3.1).
func PropertyNames[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.PropertyNames = New[T](store.Store)
	}
}

// UnevaluatedProperties sets the schema for properties not covered by properties or patternProperties (OpenAPI 3.1).
// Use UnevaluatedProperties[any]() to allow any extra properties, or a concrete type to constrain them.
func UnevaluatedProperties[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.UnevaluatedProperties = &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: New[T](store.Store),
		}
	}
}

// unions

// OneOf requires the value to be valid against exactly one of the given schemas.
//
//	OneOf(Type[A](), Type[B]()))
func OneOf(types ...TypeGetter) Option {
	return func(s *base.Schema, store *Store) {
		if len(types) == 0 {
			return
		}
		schemas := make([]*base.SchemaProxy, len(types))
		for i, t := range types {
			schemas[i] = t(store.Store)
		}
		s.OneOf = schemas
	}
}

// OneOfUnion expands a union.Union or union.TaggedUnion type T into oneOf case schemas.
func OneOfUnion[T any]() Option {
	return func(s *base.Schema, store *Store) {
		rt := reflect.TypeFor[T]()
		if info := unionFromType(rt, store.Store, make(map[reflect.Type]bool)); info != nil {
			s.OneOf = info.schemas
			if info.discriminator != nil {
				s.Discriminator = info.discriminator
			}
		}
	}
}

// AnyOf requires the value to be valid against at least one of the given schemas.
//
//	AnyOf(Type[A](), Type[B]()))
//
// If called with no arguments, the schemas are moved from s.OneOf to s.AnyOf.
func AnyOf(types ...TypeGetter) Option {
	return func(s *base.Schema, store *Store) {
		if len(types) == 0 {
			s.AnyOf = s.OneOf
			s.OneOf = nil
			return
		}
		schemas := make([]*base.SchemaProxy, len(types))
		for i, t := range types {
			schemas[i] = t(store.Store)
		}
		s.AnyOf = schemas
	}
}

// AnyOfUnion expands a union.Union or union.TaggedUnion type T into anyOf case schemas.
func AnyOfUnion[T any]() Option {
	return func(s *base.Schema, store *Store) {
		rt := reflect.TypeFor[T]()
		if info := unionFromType(rt, store.Store, make(map[reflect.Type]bool)); info != nil {
			s.AnyOf = info.schemas
			if info.discriminator != nil {
				s.Discriminator = info.discriminator
			}
		}
	}
}

// AllOf requires the value to be valid against all of the given schemas.
//
//	AllOf(Type[A](), Type[B]()))
//
// If called with no arguments, the schemas are moved from s.OneOf to s.AllOf.
func AllOf(types ...TypeGetter) Option {
	return func(s *base.Schema, store *Store) {
		if len(types) == 0 {
			s.AllOf = s.OneOf
			s.OneOf = nil
			return
		}
		schemas := make([]*base.SchemaProxy, len(types))
		for i, t := range types {
			schemas[i] = t(store.Store)
		}
		s.AllOf = schemas
	}
}

// AllOfUnion expands a union.Union or union.TaggedUnion type T into allOf case schemas.
func AllOfUnion[T any]() Option {
	return func(s *base.Schema, store *Store) {
		rt := reflect.TypeFor[T]()
		if info := unionFromType(rt, store.Store, make(map[reflect.Type]bool)); info != nil {
			s.AllOf = info.schemas
			if info.discriminator != nil {
				s.Discriminator = info.discriminator
			}
		}
	}
}

// Not requires the value to be invalid against the given schema.
func Not[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.Not = New[T](store.Store)
	}
}

// conditionals

// If sets the conditional schema — if this validates, Then is applied; otherwise Else is applied (OpenAPI 3.1).
func If[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.If = New[T](store.Store)
	}
}

// Then sets the schema applied when If validates (OpenAPI 3.1).
func Then[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.Then = New[T](store.Store)
	}
}

// Else sets the schema applied when If does not validate (OpenAPI 3.1).
func Else[T any]() Option {
	return func(s *base.Schema, store *Store) {
		s.Else = New[T](store.Store)
	}
}

// xml

// XMLName sets the XML element or attribute name, overriding the property name.
func XMLName(name string) Option {
	return func(s *base.Schema, _ *Store) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Name = name
	}
}

// XMLNamespace sets the URI of the XML namespace for the element.
func XMLNamespace(namespace string) Option {
	return func(s *base.Schema, _ *Store) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Namespace = namespace
	}
}

// XMLPrefix sets the prefix to use with the XML namespace.
func XMLPrefix(prefix string) Option {
	return func(s *base.Schema, _ *Store) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Prefix = prefix
	}
}

// XMLAttribute marks the property as an XML attribute rather than an element. Default: false.
func XMLAttribute() Option {
	return func(s *base.Schema, _ *Store) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Attribute = true
	}
}

// XMLWrapped wraps an array in an enclosing XML element. Default: false.
func XMLWrapped() Option {
	return func(s *base.Schema, _ *Store) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Wrapped = true
	}
}
