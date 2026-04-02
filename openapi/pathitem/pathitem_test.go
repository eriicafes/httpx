package pathitem

import (
	"net/http"
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/param"
	"github.com/eriicafes/httpx/openapi/server"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestPathBuilders(t *testing.T) {
	st := store.New(&v3.Components{})
	p := New(
		Summary("summary"),
		Description("description"),
		Parameter[string](param.InHeader, "X-Region", param.Description("region")),
		Server("https://example.com", "prod", server.Name("production")),
		AdditionalOperation("trace", op.Summary("trace op")),
	)
	p.Operation(http.MethodGet, op.Summary("get"))
	p.Route(http.MethodPost, op.Summary("post"), func(http.ResponseWriter, *http.Request) error { return nil })
	p.Handle(http.MethodPut, op.Summary("put"), http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	p.HandleFunc(http.MethodPatch, op.Summary("patch"), func(http.ResponseWriter, *http.Request) {})

	item := GetPathItem(p, st, nil)
	if item.Summary != "summary" || item.Description != "description" {
		t.Fatalf("GetPathItem() did not apply path options: %#v", item)
	}
	if item.Get == nil || item.Get.Summary != "get" {
		t.Fatalf("GetPathItem() did not add GET operation: %#v", item.Get)
	}
	if item.Post == nil || item.Post.Summary != "post" {
		t.Fatalf("GetPathItem() did not add POST operation: %#v", item.Post)
	}
	if item.Put == nil || item.Put.Summary != "put" {
		t.Fatalf("GetPathItem() did not add PUT operation: %#v", item.Put)
	}
	if item.Patch == nil || item.Patch.Summary != "patch" {
		t.Fatalf("GetPathItem() did not add PATCH operation: %#v", item.Patch)
	}
	if len(item.Parameters) != 1 || item.Parameters[0].Description != "region" {
		t.Fatalf("GetPathItem() did not add shared parameter: %#v", item.Parameters)
	}
	if len(item.Servers) != 1 || item.Servers[0].Name != "production" {
		t.Fatalf("GetPathItem() did not add server: %#v", item.Servers)
	}
	if item.AdditionalOperations == nil {
		t.Fatal("GetPathItem() did not initialize AdditionalOperations")
	}
	if trace, ok := item.AdditionalOperations.Get("trace"); !ok || trace.Summary != "trace op" {
		t.Fatalf("GetPathItem() did not set additional op: %#v ok=%v", trace, ok)
	}

	handlers := GetHandlers(p)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodPatch} {
		if _, ok := handlers[method]; !ok {
			t.Fatalf("GetHandlers() missing %s handler", method)
		}
	}

	built := p.PathItem()(st)
	if built.Get == nil || built.Get.Summary != "get" {
		t.Fatalf("PathItem() returned wrong builder result: %#v", built.Get)
	}
}

func TestGetPathItem_StoresReference(t *testing.T) {
	st := store.New(&v3.Components{})
	p := New(Reference("SharedPath"), Description("shared"))
	p.Operation(http.MethodGet, op.Summary("get"))

	item := GetPathItem(p, st, nil)
	if item.Reference != "#/components/pathItems/SharedPath" {
		t.Fatalf("GetPathItem() ref = %q, want component ref", item.Reference)
	}
	ref, ok := st.GetPathItem("SharedPath")
	if !ok || ref.Reference != item.Reference {
		t.Fatalf("GetPathItem() did not store referenced path item: %#v ok=%v", ref, ok)
	}
}
