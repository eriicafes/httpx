package link

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNew_AppliesOptionsAndStoresReference(t *testing.T) {
	st := store.New(&v3.Components{})

	ref := New(st,
		OperationId("getUser"),
		Parameter("id", "$response.body#/id"),
		RequestBody("$request.body"),
		Description("follow up"),
		Server("https://api.example.com", "prod"),
		Reference("UserLink"),
	)
	if ref.Reference != "#/components/links/UserLink" {
		t.Fatalf("New() ref = %q, want component ref", ref.Reference)
	}
	stored, ok := st.GetLink("UserLink")
	if !ok || stored.Reference != ref.Reference {
		t.Fatalf("New() did not store link component: %#v ok=%v", stored, ok)
	}

	inline := New(st, OperationRef("#/paths/~1users/get"))
	if inline.OperationRef != "#/paths/~1users/get" {
		t.Fatalf("New() did not apply OperationRef: %#v", inline)
	}
}
