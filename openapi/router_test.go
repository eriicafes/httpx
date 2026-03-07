package openapi

import (
	"net/http"
	"testing"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/doc"
	"github.com/eriicafes/httpx/openapi/op"
)

var normalizePathTests = []struct {
	name  string
	input string
	want  string
}{
	{
		name:  "simple path",
		input: "/users",
		want:  "/users",
	},
	{
		name:  "path with named param",
		input: "/users/{id}",
		want:  "/users/{id}",
	},
	{
		name:  "catch-all wildcard trimmed",
		input: "/files/{path...}",
		want:  "/files/{path}",
	},
	{
		name:  "exact-match marker dropped",
		input: "/users/{$}",
		want:  "/users",
	},
	{
		name:  "trailing slash becomes path param",
		input: "/files/",
		want:  "/files/{path}",
	},
	{
		name:  "root trailing slash",
		input: "/",
		want:  "/{path}",
	},
	{
		name:  "root exact match",
		input: "/{$}",
		want:  "",
	},
	{
		name:  "nested path",
		input: "/api/v1/users/{id}",
		want:  "/api/v1/users/{id}",
	},
	{
		name:  "catch-all wildcard in middle",
		input: "/a/{b...}/c",
		want:  "/a/{b}/c",
	},
}

func TestNormalizePath(t *testing.T) {
	for _, tt := range normalizePathTests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewRouter_Defaults(t *testing.T) {
	r := NewRouter()
	d := r.GetDocument()

	if d.Version != "3.1.0" {
		t.Errorf("expected version 3.1.0, got %q", d.Version)
	}
	if d.Info == nil {
		t.Fatal("expected Info to be set")
	}
	if d.Info.Title != "API" {
		t.Errorf("expected title 'API', got %q", d.Info.Title)
	}
	if d.Info.Version != "0.0.0" {
		t.Errorf("expected version '0.0.0', got %q", d.Info.Version)
	}
	if d.Paths == nil {
		t.Fatal("expected Paths to be initialized")
	}
	if d.Components == nil {
		t.Fatal("expected Components to be initialized")
	}
}

func TestNewRouter_WithOptions(t *testing.T) {
	r := NewRouter(
		doc.Info("My API", "2.0.0"),
		doc.OpenAPIVersion("3.0.3"),
	)
	d := r.GetDocument()

	if d.Info.Title != "My API" {
		t.Errorf("expected title 'My API', got %q", d.Info.Title)
	}
	if d.Info.Version != "2.0.0" {
		t.Errorf("expected version '2.0.0', got %q", d.Info.Version)
	}
	if d.Version != "3.0.3" {
		t.Errorf("expected openapi version '3.0.3', got %q", d.Version)
	}
}

func TestRouter_WithMux_SharesDocument(t *testing.T) {
	r1 := NewRouter(doc.Info("Shared", "1.0"))
	r2 := r1.WithMux(http.NewServeMux())

	if r1.GetDocument() != r2.GetDocument() {
		t.Error("expected WithMux to share the same document")
	}
}

func TestRouter_Path_GET(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("GET /users", op.Summary("list users"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected path item for /users")
	}
	if item.Get == nil {
		t.Fatal("expected GET operation to be set")
	}
	if item.Get.Summary != "list users" {
		t.Errorf("expected summary 'list users', got %q", item.Get.Summary)
	}
	if item.Post != nil {
		t.Error("expected POST to be nil for GET-only path")
	}
}

func TestRouter_Path_POST(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("POST /users", op.Summary("create user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected path item for /users")
	}
	if item.Post == nil {
		t.Fatal("expected POST operation to be set")
	}
	if item.Post.Summary != "create user" {
		t.Errorf("expected summary 'create user', got %q", item.Post.Summary)
	}
	if item.Get != nil {
		t.Error("expected GET to be nil for POST-only path")
	}
}

func TestRouter_Path_PUT(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("PUT /users/{id}", op.Summary("update user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Put == nil {
		t.Fatal("expected PUT operation to be set")
	}
}

func TestRouter_Path_PATCH(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("PATCH /users/{id}", op.Summary("patch user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Patch == nil {
		t.Fatal("expected PATCH operation to be set")
	}
}

func TestRouter_Path_DELETE(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("DELETE /users/{id}", op.Summary("delete user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Delete == nil {
		t.Fatal("expected DELETE operation to be set")
	}
}

func TestRouter_Path_NoMethod_SetsAllMethods(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("/users", op.Summary("all methods"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected path item for /users")
	}
	if item.Get == nil {
		t.Error("expected GET to be set for no-method pattern")
	}
	if item.Post == nil {
		t.Error("expected POST to be set for no-method pattern")
	}
	if item.Put == nil {
		t.Error("expected PUT to be set for no-method pattern")
	}
	if item.Patch == nil {
		t.Error("expected PATCH to be set for no-method pattern")
	}
	if item.Delete == nil {
		t.Error("expected DELETE to be set for no-method pattern")
	}
}

func TestRouter_Path_MultiplePaths(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("GET /users", op.Summary("list"))
	r.Path("POST /users", op.Summary("create"))
	r.Path("GET /items", op.Summary("items"))

	d := r.GetDocument()

	usersItem, ok := d.Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected /users path item")
	}
	if usersItem.Get == nil || usersItem.Get.Summary != "list" {
		t.Errorf("expected GET /users summary 'list'")
	}
	if usersItem.Post == nil || usersItem.Post.Summary != "create" {
		t.Errorf("expected POST /users summary 'create'")
	}

	if _, ok := d.Paths.PathItems.Get("/items"); !ok {
		t.Fatal("expected /items path item")
	}
}

func TestRouter_Path_WithMuxPrefix(t *testing.T) {
	base := http.NewServeMux()
	prefixed := httpx.Prefix(base, "/api")
	r := NewRouter().WithMux(prefixed)

	r.Path("GET /users", op.Summary("list"))

	d := r.GetDocument()
	if _, ok := d.Paths.PathItems.Get("/api/users"); !ok {
		t.Error("expected path /api/users (mux prefix should be prepended)")
	}
	if _, ok := d.Paths.PathItems.Get("/users"); ok {
		t.Error("expected path /users without prefix to not exist")
	}
}

func TestRouter_Path_NormalizesWildcard(t *testing.T) {
	r := NewRouter().WithMux(http.NewServeMux())
	r.Path("GET /files/{path...}", op.Summary("download"))

	if _, ok := r.GetDocument().Paths.PathItems.Get("/files/{path}"); !ok {
		t.Error("expected normalized path /files/{path}")
	}
}

func TestUseRouter_Found(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base, doc.Info("Test", "1.0"))

	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected UseRouter to return a non-nil Router")
	}
	if r.GetDocument().Info.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", r.GetDocument().Info.Title)
	}
}

func TestUseRouter_NotFound(t *testing.T) {
	r := UseRouter(http.NewServeMux())
	if r != nil {
		t.Error("expected UseRouter to return nil when no openapiMux is in the chain")
	}
}

func TestUseRouter_ThroughPrefix(t *testing.T) {
	base := http.NewServeMux()
	withRouter := WithRouter(base, doc.Info("Prefixed", "1.0"))
	prefixed := httpx.Prefix(withRouter, "/api")

	r := UseRouter(prefixed)
	if r == nil {
		t.Fatal("expected UseRouter to find Router through Prefix layer")
	}
	if r.GetDocument().Info.Title != "Prefixed" {
		t.Errorf("expected title 'Prefixed', got %q", r.GetDocument().Info.Title)
	}
}

func TestUseRouter_SharesDocument(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base)

	r1 := UseRouter(wrapped)
	r2 := UseRouter(wrapped)

	if r1 == nil || r2 == nil {
		t.Fatal("expected non-nil routers")
	}
	if r1.GetDocument() != r2.GetDocument() {
		t.Error("expected UseRouter to share the same document across calls")
	}
}

func TestUseRouter_BindsOriginalMux(t *testing.T) {
	// UseRouter binds the returned Router to the mux passed in (not the inner one).
	// Operations registered via the Router should use that mux's prefix.
	base := http.NewServeMux()
	withRouter := WithRouter(base, doc.Info("API", "1.0"))
	prefixed := httpx.Prefix(withRouter, "/v1")

	r := UseRouter(prefixed)
	if r == nil {
		t.Fatal("expected non-nil router")
	}

	r.Path("GET /users", op.Summary("list"))

	if _, ok := r.GetDocument().Paths.PathItems.Get("/v1/users"); !ok {
		t.Error("expected path /v1/users using the bound mux prefix")
	}
}
