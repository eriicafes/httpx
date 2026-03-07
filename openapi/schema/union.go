package schema

import (
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	"github.com/pb33f/libopenapi/orderedmap"
)

const unionPkg = "github.com/eriicafes/union"

type unionInfo struct {
	schemas       []*base.SchemaProxy
	discriminator *base.Discriminator
}

// unionFromType checks if t is a union.Union or union.TaggedUnion and returns
// the expanded case schemas with an optional discriminator. Returns nil if t is
// not a union type.
func unionFromType(t reflect.Type) *unionInfo {
	if t.PkgPath() != unionPkg {
		return nil
	}

	name := t.Name()
	isTagged := len(name) > 11 && name[:11] == "TaggedUnion"
	isUntagged := len(name) > 5 && name[:5] == "Union"
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
		// Resolve field names via a pointer to a zero Spec value so that both
		// value and pointer receivers on TaggedFieldNames are satisfied.
		variantKey, valueKey := "type", "value"
		if spec, ok := reflect.New(specType).Interface().(interface {
			TaggedFieldNames() (string, string)
		}); ok {
			variantKey, valueKey = spec.TaggedFieldNames()
		}
		return taggedUnionFromSpec(specType, variantKey, valueKey)
	}
	return untaggedUnionFromSpec(specType)
}

func untaggedUnionFromSpec(specType reflect.Type) *unionInfo {
	info := &unionInfo{}
	for i := 0; i < specType.NumField(); i++ {
		f := specType.Field(i)
		info.schemas = append(info.schemas, base.CreateSchemaProxy(SchemaFromType(f.Type)))
	}
	return info
}

// taggedUnionFromSpec builds case schemas for a TaggedUnion.
// Each case is an object with a const discriminator property (variantKey) and a
// value property (valueKey) holding the field's schema, matching the wire format:
//
//	{ "<variantKey>": "<variantName>", "<valueKey>": { ... } }
func taggedUnionFromSpec(specType reflect.Type, variantKey, valueKey string) *unionInfo {
	info := &unionInfo{
		discriminator: &base.Discriminator{PropertyName: variantKey},
	}

	for i := 0; i < specType.NumField(); i++ {
		f := specType.Field(i)
		variantName := f.Tag.Get("variant")
		if variantName == "" {
			variantName = f.Name
		}

		props := orderedmap.New[string, *base.SchemaProxy]()
		props.Set(variantKey, base.CreateSchemaProxy(&base.Schema{
			Const: toYAMLNode(variantName),
		}))
		props.Set(valueKey, base.CreateSchemaProxy(SchemaFromType(f.Type)))

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
