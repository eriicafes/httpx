package param

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/op/example"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type parameterDefaults string

func (parameterDefaults) Parameter() Option {
	return Description("from type")
}

type parameterReference string

func (parameterReference) Parameter() Option {
	return Options(
		Reference("StatusParam"),
		Description("from type"),
	)
}

func TestNew_AppliesOptions(t *testing.T) {
	st := store.New(&v3.Components{})

	inline := New[string](InHeader, "X-Request-Id", st,
		Description("request id"),
		Required(true),
		Deprecated(),
		AllowEmptyValue(),
		Style("simple"),
		Explode(true),
		AllowReserved(),
		Example("req_123"),
		NamedExample("basic", "req_123", example.Summary("basic")),
		NamedExampleRef("stored", "StoredExample"),
	)
	if inline.Description != "request id" || inline.Style != "simple" {
		t.Fatalf("New() basic options incorrect: description=%q style=%q", inline.Description, inline.Style)
	}
	if inline.Required == nil || !*inline.Required || !inline.Deprecated || !inline.AllowEmptyValue || !inline.AllowReserved {
		t.Fatalf("New() bool flags incorrect: required=%v deprecated=%t allowEmpty=%t allowReserved=%t", inline.Required, inline.Deprecated, inline.AllowEmptyValue, inline.AllowReserved)
	}
	if inline.Examples == nil || inline.Examples.Len() != 2 {
		t.Fatalf("New() did not set named examples: count=%d", inline.Examples.Len())
	}
	basic, _ := inline.Examples.Get("basic")
	if basic == nil || basic.Summary != "basic" {
		t.Fatalf("NamedExample() incorrect: nil=%t summary=%q", basic == nil, func() string { if basic == nil { return "" }; return basic.Summary }())
	}
	stored, _ := inline.Examples.Get("stored")
	if stored == nil || stored.Reference != "#/components/examples/StoredExample" {
		t.Fatalf("NamedExampleRef() incorrect: nil=%t ref=%q", stored == nil, func() string { if stored == nil { return "" }; return stored.Reference }())
	}
}

func TestNew_AppliesTypeDefaults(t *testing.T) {
	st := store.New(&v3.Components{})

	ref := New[parameterDefaults](InQuery, "status", st)
	if ref.Description != "from type" {
		t.Fatalf("type defaults not applied: description=%q", ref.Description)
	}
}

func TestNew_TypeLevelReferenceOmitsInlineOptions(t *testing.T) {
	components := &v3.Components{}
	st := store.New(components)

	ref := New[parameterReference](InQuery, "status", st, Required(true))
	if ref.Reference != "#/components/parameters/StatusParam" {
		t.Fatalf("New() ref = %q, want component ref", ref.Reference)
	}

	component, ok := components.Parameters.Get("StatusParam")
	if !ok {
		t.Fatal("expected stored referenced parameter")
	}
	if component.Description != "from type" {
		t.Fatalf("type defaults not preserved on stored component: description=%q", component.Description)
	}
	if component.Required != nil {
		t.Fatalf("inline options should be skipped when type sets reference: required=%v", component.Required)
	}
}
