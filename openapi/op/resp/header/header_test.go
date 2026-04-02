package header

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type headerDefaults string

func (headerDefaults) Header() Option {
	return Description("from type")
}

type headerReference string

func (headerReference) Header() Option {
	return Options(
		Reference("RequestID"),
		Description("from type"),
	)
}

func TestNew_AppliesOptions(t *testing.T) {
	st := store.New(&v3.Components{})

	inline := New[string](st,
		Description("trace id"),
		Required(true),
		Deprecated(),
		AllowEmptyValue(),
		Style("simple"),
		Explode(),
		AllowReserved(),
		Example("req_123"),
		NamedExample("basic", "req_456", example.Summary("basic")),
		NamedExampleRef("stored", "StoredExample"),
	)
	if inline.Description != "trace id" || inline.Style != "simple" {
		t.Fatalf("New() basic options incorrect: description=%q style=%q", inline.Description, inline.Style)
	}
	if !inline.Required || !inline.Deprecated || !inline.AllowEmptyValue || !inline.Explode || !inline.AllowReserved {
		t.Fatalf("New() flags incorrect: required=%t deprecated=%t allowEmpty=%t explode=%t allowReserved=%t", inline.Required, inline.Deprecated, inline.AllowEmptyValue, inline.Explode, inline.AllowReserved)
	}
	if inline.Examples == nil || inline.Examples.Len() != 2 {
		t.Fatalf("New() did not populate examples: count=%d", inline.Examples.Len())
	}
}

func TestNew_AppliesTypeDefaults(t *testing.T) {
	st := store.New(&v3.Components{})

	ref := New[headerDefaults](st)
	if ref.Description != "from type" {
		t.Fatalf("type defaults not applied: description=%q", ref.Description)
	}
}

func TestNew_TypeLevelReferenceOmitsInlineOptions(t *testing.T) {
	components := &v3.Components{}
	st := store.New(components)

	ref := New[headerReference](st, Required(true))
	if ref.Reference != "#/components/headers/RequestID" {
		t.Fatalf("New() ref = %q, want component ref", ref.Reference)
	}

	component, ok := components.Headers.Get("RequestID")
	if !ok {
		t.Fatal("expected stored referenced header")
	}
	if component.Description != "from type" {
		t.Fatalf("type defaults not preserved on stored component: description=%q", component.Description)
	}
	if component.Required {
		t.Fatalf("inline options should be skipped when type sets reference: required=%t", component.Required)
	}
}
