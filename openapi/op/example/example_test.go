package example

import (
	"testing"

	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNew_AndNewRef(t *testing.T) {
	st := store.New(&v3.Components{})

	ex := New(st, map[string]any{"status": "ok"}, Summary("summary"), Description("description"), Reference("HealthExample"))
	if ex.Reference != "#/components/examples/HealthExample" {
		t.Fatalf("New() ref = %q, want component ref", ex.Reference)
	}
	ref, ok := st.GetExample("HealthExample")
	if !ok || ref.Reference != ex.Reference {
		t.Fatalf("New() did not store example component: %#v ok=%v", ref, ok)
	}

	inline := New(st, "ok", ExternalValue("https://example.com/example.txt"))
	if inline.Value == nil || inline.ExternalValue != "https://example.com/example.txt" {
		t.Fatalf("New() did not apply example options: %#v", inline)
	}

	if got := NewRef("OtherExample").Reference; got != "#/components/examples/OtherExample" {
		t.Fatalf("NewRef() = %q, want component ref", got)
	}
}
