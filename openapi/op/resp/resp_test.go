package resp_test

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/mediatype"
	"github.com/eriicafes/httpx/openapi/op/resp"
	"github.com/eriicafes/httpx/openapi/op/resp/header"
	"github.com/eriicafes/httpx/openapi/op/resp/link"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type responseDefaults struct{}

func (responseDefaults) Response() resp.Option {
	return resp.Description("from type")
}

type responseReference struct{}

func (responseReference) Response() resp.Option {
	return resp.Options(resp.Description("from type"), resp.Reference("UserResponse"))
}

type responseMediaDefaults struct{}

func (responseMediaDefaults) MediaType() mediatype.Option {
	return mediatype.Example("media example")
}

func TestNew_AppliesOptions(t *testing.T) {
	st := store.New(&v3.Components{})

	inline := resp.New[responseMediaDefaults](st,
		resp.Summary("summary"),
		resp.Description("description"),
		resp.Header[string]("X-Request-Id", header.Description("trace id")),
		resp.Link("user", link.OperationId("getUser")),
		resp.ContentType[responseMediaDefaults]("application/xml"),
	)
	if inline.Summary != "summary" || inline.Description != "description" {
		t.Fatalf("New() summary=%q description=%q, want %q and %q", inline.Summary, inline.Description, "summary", "description")
	}
	if inline.Headers == nil || inline.Links == nil || inline.Content == nil {
		t.Fatalf("New() headers nil=%t links nil=%t content nil=%t, want all initialized", inline.Headers == nil, inline.Links == nil, inline.Content == nil)
	}
	jsonMedia, ok := inline.Content.Get("application/json")
	if !ok || jsonMedia.Example == nil {
		t.Fatalf("default media type missing type defaults: hasJSON=%t hasExample=%t", ok, ok && jsonMedia.Example != nil)
	}
	if _, ok := inline.Content.Get("application/xml"); !ok {
		t.Fatal("ContentType() did not add XML content")
	}

	noContent := resp.New[op.NoContent](st, resp.Description("empty"))
	if noContent.Content.Len() != 0 {
		t.Fatalf("New[op.NoContent]() should not include content: len=%d", noContent.Content.Len())
	}
}

func TestNew_AppliesTypeDefaults(t *testing.T) {
	st := store.New(&v3.Components{})

	response := resp.New[responseDefaults](st)
	if response.Description != "from type" {
		t.Fatalf("type defaults not applied: description=%q", response.Description)
	}
}

func TestNew_TypeLevelReferenceOmitsInlineOptions(t *testing.T) {
	components := &v3.Components{}
	st := store.New(components)

	response := resp.New[responseReference](st,
		resp.Summary("user response"),
		resp.Description("returned user"),
	)
	if response.Reference != "#/components/responses/UserResponse" {
		t.Fatalf("New() ref = %q, want component ref", response.Reference)
	}

	stored, ok := st.GetResponse("UserResponse")
	if !ok || stored.Reference != response.Reference {
		t.Fatalf("New() did not store response component: ok=%t ref=%q want=%q", ok, stored.Reference, response.Reference)
	}

	component, ok := components.Responses.Get("UserResponse")
	if !ok {
		t.Fatal("expected stored referenced response component")
	}
	if component.Description != "from type" {
		t.Fatalf("type defaults not preserved on stored component: description=%q", component.Description)
	}
	if component.Summary != "" {
		t.Fatalf("inline options should be skipped when type sets reference: summary=%q", component.Summary)
	}
}
