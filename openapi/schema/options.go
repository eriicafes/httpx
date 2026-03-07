package schema

import (
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
	"go.yaml.in/yaml/v4"
)

// Option configures a JSON Schema and its associated RegistryField metadata.
type Option func(*base.Schema, *RegistryField)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(s *base.Schema, f *RegistryField) {
		for _, opt := range opts {
			opt(s, f)
		}
	}
}

// Ref registers the type in components/schemas under name and returns a $ref to it.
func Ref(name string) Option {
	return func(_ *base.Schema, f *RegistryField) {
		f.Reference = name
	}
}

// general

// Title sets the schema title.
func Title(s string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Title = s
	}
}

// Description sets the schema description.
func Description(s string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Description = s
	}
}

// Format sets the format hint for the schema (e.g. "date-time", "email", "uuid").
// Prefer the named helpers (Email, UUID, DateTime) for common formats.
func Format(s string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Format = s
	}
}

// ReadOnly marks the schema as read-only — the value must not be sent in requests. Default: false.
func ReadOnly() Option {
	v := true
	return func(schema *base.Schema, _ *RegistryField) {
		schema.ReadOnly = &v
	}
}

// WriteOnly marks the schema as write-only — the value must not be included in responses. Default: false.
func WriteOnly() Option {
	v := true
	return func(schema *base.Schema, _ *RegistryField) {
		schema.WriteOnly = &v
	}
}

// Deprecated marks the schema as deprecated. Default: false.
func Deprecated() Option {
	v := true
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Deprecated = &v
	}
}

// Nullable marks the schema as nullable (OpenAPI 3.0 compatibility). Default: false.
// In OpenAPI 3.1 prefer using a pointer type, which sets nullable automatically.
func Nullable() Option {
	v := true
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Nullable = &v
	}
}

// ExternalDocs attaches an external documentation link to the schema.
func ExternalDocs(url, description string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Default sets the default value used when the field is absent.
func Default(v any) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Default = toYAMLNode(v)
	}
}

// Example sets an example value for the schema.
func Example(v any) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Example = toYAMLNode(v)
	}
}

// Examples sets the array of example values for the schema (OpenAPI 3.1).
func Examples(values ...any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		nodes := make([]*yaml.Node, len(values))
		for i, v := range values {
			nodes[i] = toYAMLNode(v)
		}
		s.Examples = nodes
	}
}

// Const restricts the schema to a single constant value.
func Const(v any) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.Const = toYAMLNode(v)
	}
}

// Enum restricts the schema to a fixed set of allowed values.
func Enum(values ...any) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		nodes := make([]*yaml.Node, len(values))
		for i, v := range values {
			nodes[i] = toYAMLNode(v)
		}
		schema.Enum = nodes
	}
}

// string

// MinLength sets the minimum number of characters for a string value.
func MinLength(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MinLength = &n
	}
}

// MaxLength sets the maximum number of characters for a string value.
func MaxLength(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MaxLength = &n
	}
}

// Pattern restricts a string value to match the given regular expression.
func Pattern(p string) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Pattern = p
	}
}

// Email sets the format to "email".
func Email() Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Format = "email"
	}
}

// UUID sets the format to "uuid".
func UUID() Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Format = "uuid"
	}
}

// DateTime sets the format to "date-time".
func DateTime() Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Format = "date-time"
	}
}

// ContentEncoding sets the encoding for string content (e.g. "base64", "quoted-printable").
func ContentEncoding(s string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.ContentEncoding = s
	}
}

// ContentMediaType sets the MIME type of string content (e.g. "image/png", "application/pdf").
func ContentMediaType(s string) Option {
	return func(schema *base.Schema, _ *RegistryField) {
		schema.ContentMediaType = s
	}
}

// numeric

// MultipleOf requires the numeric value to be a multiple of the given number.
func MultipleOf(v float64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MultipleOf = &v
	}
}

// Minimum sets the inclusive lower bound for a numeric value.
func Minimum(v float64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Minimum = &v
	}
}

// Maximum sets the inclusive upper bound for a numeric value.
func Maximum(v float64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Maximum = &v
	}
}

// ExclusiveMinimum sets the exclusive lower bound for a numeric value (OpenAPI 3.1).
func ExclusiveMinimum(v float64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.ExclusiveMinimum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	}
}

// ExclusiveMaximum sets the exclusive upper bound for a numeric value (OpenAPI 3.1).
func ExclusiveMaximum(v float64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.ExclusiveMaximum = &base.DynamicValue[bool, float64]{N: 1, B: v}
	}
}

// array

// MinItems sets the minimum number of elements in an array.
func MinItems(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MinItems = &n
	}
}

// MaxItems sets the maximum number of elements in an array.
func MaxItems(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MaxItems = &n
	}
}

// UniqueItems requires all elements in an array to be distinct. Default: false.
func UniqueItems() Option {
	v := true
	return func(s *base.Schema, _ *RegistryField) {
		s.UniqueItems = &v
	}
}

// MinContains sets the minimum number of items that must validate against the contains schema.
func MinContains(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MinContains = &n
	}
}

// MaxContains sets the maximum number of items that may validate against the contains schema.
func MaxContains(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MaxContains = &n
	}
}

// PrefixItems sets the tuple validation schemas for each position in an array (OpenAPI 3.1).
// Pass zero values of each positional type (e.g. PrefixItems(int(0), "")).
func PrefixItems(types ...any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		schemas := make([]*base.SchemaProxy, len(types))
		for i, t := range types {
			schemas[i] = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
		}
		s.PrefixItems = schemas
	}
}

// Contains sets the schema that at least one item in the array must validate against (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. Contains(MyStruct{})).
func Contains(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Contains = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// UnevaluatedItems sets the schema for array items not covered by items or prefixItems (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. UnevaluatedItems(MyStruct{})).
func UnevaluatedItems(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.UnevaluatedItems = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// object

// MinProperties sets the minimum number of properties an object must have.
func MinProperties(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MinProperties = &n
	}
}

// MaxProperties sets the maximum number of properties an object may have.
func MaxProperties(n int64) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.MaxProperties = &n
	}
}

// AdditionalProperties overrides the default closed-object behavior for structs.
// Use AdditionalProperties[any]() to allow any extra fields, or a concrete type
// to constrain their schema.
func AdditionalProperties[T any]() Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.AdditionalProperties = &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(SchemaFromType(reflect.TypeFor[T]())),
		}
	}
}

// Field applies options to a named property of a struct schema.
// If the property does not exist the option is a no-op.
func Field(name string, opts ...Option) Option {
	return func(s *base.Schema, _ *RegistryField) {
		prop, ok := s.Properties.Get(name)
		if !ok {
			return
		}
		f := &RegistryField{}
		for _, opt := range opts {
			opt(prop.Schema(), f)
		}
	}
}

// PatternProperties adds schema constraints for properties whose names match the given regex.
func PatternProperties(pattern string, opts ...Option) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.PatternProperties == nil {
			s.PatternProperties = orderedmap.New[string, *base.SchemaProxy]()
		}
		inner := &base.Schema{}
		f := &RegistryField{}
		for _, opt := range opts {
			opt(inner, f)
		}
		s.PatternProperties.Set(pattern, base.CreateSchemaProxy(inner))
	}
}

// DependentRequired specifies properties that are required when the named property is present (OpenAPI 3.1).
func DependentRequired(property string, required ...string) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.DependentRequired == nil {
			s.DependentRequired = orderedmap.New[string, []string]()
		}
		s.DependentRequired.Set(property, required)
	}
}

// DependentSchemas sets the schema that must validate when the named property is present (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. DependentSchemas("flag", Extra{})).
func DependentSchemas(property string, t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.DependentSchemas == nil {
			s.DependentSchemas = orderedmap.New[string, *base.SchemaProxy]()
		}
		s.DependentSchemas.Set(property, base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t))))
	}
}

// PropertyNames sets constraints on the names of object properties (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. PropertyNames("")).
func PropertyNames(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.PropertyNames = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// UnevaluatedProperties sets the schema for properties not covered by properties or patternProperties (OpenAPI 3.1).
// Use UnevaluatedProperties[any]() to allow any extra properties, or a concrete type to constrain them.
func UnevaluatedProperties[T any]() Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.UnevaluatedProperties = &base.DynamicValue[*base.SchemaProxy, bool]{
			N: 0,
			A: base.CreateSchemaProxy(SchemaFromType(reflect.TypeFor[T]())),
		}
	}
}

// unions

// OneOf requires the value to be valid against exactly one of the given schemas.
// Pass zero values of the desired types (e.g. OneOf(TypeA{}, TypeB{})).
// If a union.Union or union.TaggedUnion value is passed, its cases are expanded inline.
//
// If called with no arguments, it is a no-op (s.OneOf is already the union default).
func OneOf(types ...any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if len(types) == 0 {
			return
		}
		var schemas []*base.SchemaProxy
		for _, t := range types {
			rt := reflect.TypeOf(t)
			if info := unionFromType(rt); info != nil {
				schemas = append(schemas, info.schemas...)
				if info.discriminator != nil {
					s.Discriminator = info.discriminator
				}
			} else {
				schemas = append(schemas, base.CreateSchemaProxy(SchemaFromType(rt)))
			}
		}
		s.OneOf = schemas
	}
}

// AnyOf requires the value to be valid against at least one of the given schemas.
// Pass zero values of the desired types (e.g. AnyOf(TypeA{}, TypeB{})).
// If a union.Union or union.TaggedUnion value is passed, its cases are expanded inline.
//
// If called with no arguments, the schemas are inferred from s.OneOf (set by union type
// reflection) and moved to s.AnyOf, replacing s.OneOf.
func AnyOf(types ...any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if len(types) == 0 {
			s.AnyOf = s.OneOf
			s.OneOf = nil
			return
		}
		var schemas []*base.SchemaProxy
		for _, t := range types {
			rt := reflect.TypeOf(t)
			if info := unionFromType(rt); info != nil {
				schemas = append(schemas, info.schemas...)
				if info.discriminator != nil {
					s.Discriminator = info.discriminator
				}
			} else {
				schemas = append(schemas, base.CreateSchemaProxy(SchemaFromType(rt)))
			}
		}
		s.AnyOf = schemas
	}
}

// AllOf requires the value to be valid against all of the given schemas.
// Pass zero values of the desired types (e.g. AllOf(MyStruct{}, OtherStruct{})).
// If a union.Union or union.TaggedUnion value is passed, its cases are expanded inline.
//
// If called with no arguments, the schemas are inferred from s.OneOf (set by union type
// reflection) and moved to s.AllOf, replacing s.OneOf.
func AllOf(types ...any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if len(types) == 0 {
			s.AllOf = s.OneOf
			s.OneOf = nil
			return
		}
		var schemas []*base.SchemaProxy
		for _, t := range types {
			rt := reflect.TypeOf(t)
			if info := unionFromType(rt); info != nil {
				schemas = append(schemas, info.schemas...)
				if info.discriminator != nil {
					s.Discriminator = info.discriminator
				}
			} else {
				schemas = append(schemas, base.CreateSchemaProxy(SchemaFromType(rt)))
			}
		}
		s.AllOf = schemas
	}
}

// Not requires the value to be invalid against the given schema.
// Pass a zero value of the desired type (e.g. Not(MyStruct{})).
func Not(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Not = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// conditionals

// If sets the conditional schema — if this validates, Then is applied; otherwise Else is applied (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. If(MyStruct{})).
func If(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.If = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// Then sets the schema applied when If validates (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. Then(MyStruct{})).
func Then(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Then = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// Else sets the schema applied when If does not validate (OpenAPI 3.1).
// Pass a zero value of the desired type (e.g. Else(MyStruct{})).
func Else(t any) Option {
	return func(s *base.Schema, _ *RegistryField) {
		s.Else = base.CreateSchemaProxy(SchemaFromType(reflect.TypeOf(t)))
	}
}

// xml

// XMLName sets the XML element or attribute name, overriding the property name.
func XMLName(name string) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Name = name
	}
}

// XMLNamespace sets the URI of the XML namespace for the element.
func XMLNamespace(namespace string) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Namespace = namespace
	}
}

// XMLPrefix sets the prefix to use with the XML namespace.
func XMLPrefix(prefix string) Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Prefix = prefix
	}
}

// XMLAttribute marks the property as an XML attribute rather than an element. Default: false.
func XMLAttribute() Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Attribute = true
	}
}

// XMLWrapped wraps an array in an enclosing XML element. Default: false.
func XMLWrapped() Option {
	return func(s *base.Schema, _ *RegistryField) {
		if s.XML == nil {
			s.XML = &base.XML{}
		}
		s.XML.Wrapped = true
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
