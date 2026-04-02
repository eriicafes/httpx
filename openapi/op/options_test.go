package op_test

import (
	"net/http"
	"testing"

	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/callback"
	"github.com/eriicafes/httpx/openapi/op/param"
	"github.com/eriicafes/httpx/openapi/op/resp"
	"github.com/eriicafes/httpx/openapi/pathitem"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func TestOperationHelpers(t *testing.T) {
	opn := &v3.Operation{}
	st := store.New(&v3.Components{})

	cbPath := pathitem.New()
	cbPath.Operation(http.MethodPost, op.Summary("callback"))

	op.Options(
		op.Summary("summary"),
		op.Description("description"),
		op.OperationId("id"),
		op.Tags("users", "admin"),
		op.Deprecated(),
		op.ExternalDocs("https://docs.example.com", "docs"),
		op.Server("https://api.example.com", "prod"),
		op.Security("bearerAuth", "read"),
		op.Security(""),
		op.PathParam[int]("id"),
		op.QueryParam[string]("q", param.Description("query")),
		op.HeaderParam[string]("X-Key"),
		op.CookieParam[string]("session"),
		op.RequestBody[string](),
		op.Response[string](http.StatusCreated, resp.Description("created")),
		op.Response[op.NoContent](http.StatusNoContent),
		op.Callback("notify", callback.Expression("{$request.query.callback}", cbPath.PathItem())),
	)(opn, st)

	if opn.Summary != "summary" || opn.OperationId != "id" {
		t.Fatalf("Options() basic fields incorrect: summary=%q operationId=%q", opn.Summary, opn.OperationId)
	}
	if len(opn.Tags) != 2 || opn.Tags[1] != "admin" {
		t.Fatalf("Tags() incorrect: tags=%v", opn.Tags)
	}
	if opn.Deprecated == nil || !*opn.Deprecated {
		t.Fatal("Deprecated() did not set deprecated")
	}
	if len(opn.Servers) != 1 || opn.Servers[0].URL != "https://api.example.com" {
		t.Fatalf("Server() incorrect: count=%d url=%q", len(opn.Servers), opn.Servers[0].URL)
	}
	if len(opn.Security) != 2 || !opn.Security[1].ContainsEmptyRequirement {
		t.Fatalf("Security() incorrect: count=%d optional=%t", len(opn.Security), len(opn.Security) > 1 && opn.Security[1].ContainsEmptyRequirement)
	}
	if len(opn.Parameters) != 4 {
		t.Fatalf("expected 4 parameters, got %d", len(opn.Parameters))
	}
	if opn.Parameters[0].Required == nil || !*opn.Parameters[0].Required {
		t.Fatalf("PathParam() did not force required: required=%v", opn.Parameters[0].Required)
	}
	if opn.RequestBody == nil || opn.RequestBody.Content == nil {
		t.Fatal("RequestBody() did not set request body")
	}
	if opn.Responses == nil || opn.Responses.Codes == nil {
		t.Fatal("Response() did not initialize responses")
	}
	created, _ := opn.Responses.Codes.Get("201")
	if created == nil || created.Description != "created" {
		t.Fatalf("Response() did not set 201 response: nil=%t description=%q", created == nil, func() string { if created == nil { return "" }; return created.Description }())
	}
	noContent, _ := opn.Responses.Codes.Get("204")
	if noContent == nil || noContent.Content.Len() != 0 {
		t.Fatalf("Response[NoContent]() should not include content: nil=%t contentLen=%d", noContent == nil, func() int { if noContent == nil || noContent.Content == nil { return -1 }; return noContent.Content.Len() }())
	}
	if opn.Callbacks == nil {
		t.Fatal("Callback() did not initialize callbacks")
	}
	cb, ok := opn.Callbacks.Get("notify")
	if !ok || cb.Expression.Len() != 1 {
		t.Fatalf("Callback() did not add callback: ok=%t expressionLen=%d", ok, func() int { if cb == nil || cb.Expression == nil { return -1 }; return cb.Expression.Len() }())
	}
}
