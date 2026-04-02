package callback_test

import (
	"net/http"
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/callback"
	"github.com/eriicafes/httpx/openapi/pathitem"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestNew_ExpressionAndReference(t *testing.T) {
	st := store.New(&v3.Components{})
	p := pathitem.New()
	p.Operation(http.MethodPost, op.Summary("deliver"))

	cb := callback.New(st,
		callback.Expression("{$request.query.callback}", p.PathItem()),
		callback.Reference("StatusChanged"),
	)
	if cb.Reference != "#/components/callbacks/StatusChanged" {
		t.Fatalf("New() ref = %q, want component ref", cb.Reference)
	}
	ref, ok := st.GetCallback("StatusChanged")
	if !ok || ref.Reference != cb.Reference {
		t.Fatalf("New() did not store callback component: %#v ok=%v", ref, ok)
	}

	inline := callback.New(st, callback.Expression("{$url}", p.PathItem()))
	if inline.Expression == nil || inline.Expression.Len() != 1 {
		t.Fatalf("Expression() did not add path item: %#v", inline.Expression)
	}
}
