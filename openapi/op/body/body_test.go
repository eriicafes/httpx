package body_test

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/body"
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type requestBodyDefaults struct{}

func (requestBodyDefaults) RequestBody() body.Option {
	return body.Description("from type")
}

type requestBodyReference struct{}

func (requestBodyReference) RequestBody() body.Option {
	return body.Options(
		body.Reference("DefaultBody"),
		body.Description("from type"),
	)
}

type bodyMediaDefaults struct{}

func (bodyMediaDefaults) MediaType() mediatype.Option {
	return mediatype.Example("type example")
}

func TestNew_AppliesOptions(t *testing.T) {
	st := store.New(&v3.Components{})

	inline := body.New[bodyMediaDefaults](st,
		body.Description("inline"),
		body.Required(false),
		body.ContentType[bodyMediaDefaults]("application/xml"),
	)
	if inline.Description != "inline" {
		t.Fatalf("Description() not applied: description=%q", inline.Description)
	}
	if inline.Required == nil || *inline.Required {
		t.Fatalf("Required(false) not applied: required=%v", inline.Required)
	}
	jsonMedia, ok := inline.Content.Get("application/json")
	if !ok || jsonMedia.Example == nil {
		t.Fatalf("default application/json content missing or no type defaults: hasJSON=%t hasExample=%t", ok, ok && jsonMedia.Example != nil)
	}
	if _, ok := inline.Content.Get("application/xml"); !ok {
		t.Fatal("ContentType() did not add XML media type")
	}

	noContent := body.New[op.NoContent](st)
	if noContent.Content.Len() != 0 {
		t.Fatalf("New[op.NoContent]() should not add content: len=%d", noContent.Content.Len())
	}
}

func TestNew_AppliesTypeDefaults(t *testing.T) {
	st := store.New(&v3.Components{})

	reqBody := body.New[requestBodyDefaults](st)
	if reqBody.Description != "from type" {
		t.Fatalf("type defaults not applied: description=%q", reqBody.Description)
	}
}

func TestNew_TypeLevelReferenceOmitsInlineOptions(t *testing.T) {
	components := &v3.Components{}
	st := store.New(components)

	reqBody := body.New[requestBodyReference](st, body.Required(false))
	if reqBody.Reference != "#/components/requestBodies/DefaultBody" {
		t.Fatalf("New() ref = %q, want component ref", reqBody.Reference)
	}

	component, ok := components.RequestBodies.Get("DefaultBody")
	if !ok {
		t.Fatal("expected stored referenced request body")
	}
	if component.Description != "from type" {
		t.Fatalf("type defaults not preserved on stored component: description=%q", component.Description)
	}
	if component.Required == nil || !*component.Required {
		t.Fatalf("inline options should be skipped when type sets reference: required=%v", component.Required)
	}
}
