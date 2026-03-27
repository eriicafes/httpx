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
		want:  "/",
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
	r := NewRouter("API", "0.0.0")
	d := r.GetDocument()

	if d.Version != "3.1.0" {
		t.Errorf("expected version 3.1.0, got %q", d.Version)
	}
	if d.Info == nil {
		t.Fatal("expected Info to be set")
	}
	if d.Paths == nil {
		t.Fatal("expected Paths to be initialized")
	}
	if d.Components == nil {
		t.Fatal("expected Components to be initialized")
	}
}

func TestNewRouter_WithOptions(t *testing.T) {
	r := NewRouter("API", "0.0.0", doc.Version("3.1.1"), doc.Summary("API Summary"))
	d := r.GetDocument()

	if d.Info.Title != "API" {
		t.Errorf("expected title 'API', got %q", d.Info.Title)
	}
	if d.Info.Version != "0.0.0" {
		t.Errorf("expected version '0.0.0', got %q", d.Info.Version)
	}
	if d.Info.Summary != "API Summary" {
		t.Errorf("expected openapi version 'API Summary', got %q", d.Info.Summary)
	}
	if d.Version != "3.1.1" {
		t.Errorf("expected openapi version '3.1.1', got %q", d.Version)
	}
}

func TestRouter_WithMux_CopySharesDocument(t *testing.T) {
	r1 := NewRouter("Shared", "1.0")
	r2 := r1.WithMux(http.NewServeMux())

	if r1 == r2 {
		t.Error("expected WithMux to return different instance")
	}
	if r1.GetDocument() != r2.GetDocument() {
		t.Error("expected WithMux to share the same document")
	}
}

func TestRouter_Path_GET(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("GET /users", op.Summary("list users"))

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
}

func TestRouter_Path_POST(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("POST /users", op.Summary("create user"))

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
}

func TestRouter_Path_PUT(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("PUT /users/{id}", op.Summary("update user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Put == nil {
		t.Fatal("expected PUT operation to be set")
	}
	if item.Put.Summary != "update user" {
		t.Errorf("expected summary 'update user', got %q", item.Put.Summary)
	}
}

func TestRouter_Path_PATCH(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("PATCH /users/{id}", op.Summary("patch user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Patch == nil {
		t.Fatal("expected PATCH operation to be set")
	}
	if item.Patch.Summary != "patch user" {
		t.Errorf("expected summary 'patch user', got %q", item.Patch.Summary)
	}
}

func TestRouter_Path_DELETE(t *testing.T) {
	r := NewRouter("API", "0.0.0").WithMux(http.NewServeMux())
	r.Operation("DELETE /users/{id}", op.Summary("delete user"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users/{id}")
	if !ok {
		t.Fatal("expected path item for /users/{id}")
	}
	if item.Delete == nil {
		t.Fatal("expected DELETE operation to be set")
	}
	if item.Delete.Summary != "delete user" {
		t.Errorf("expected summary 'delete user', got %q", item.Delete.Summary)
	}
}

func TestRouter_Path_NoMethod(t *testing.T) {
	r := NewRouter("API", "0.0.0").WithMux(http.NewServeMux())
	r.Operation("/users", op.Summary("all methods"))

	item, ok := r.GetDocument().Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected path item for /users")
	}
	getOp := item.Get
	if getOp == nil {
		t.Error("expected GET to be set for no-method pattern")
	}
	if item.Post != getOp {
		t.Error("expected POST to be set for no-method pattern")
	}
	if item.Put != getOp {
		t.Error("expected PUT to be set for no-method pattern")
	}
	if item.Patch != getOp {
		t.Error("expected PATCH to be set for no-method pattern")
	}
	if item.Delete != getOp {
		t.Error("expected DELETE to be set for no-method pattern")
	}
}

func TestRouter_Path_MultiplePaths(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("GET /users", op.Summary("list"))
	r.Operation("POST /users", op.Summary("create"))
	r.Operation("PUT /posts/{id}", op.Summary("update"))

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

	postsItem, ok := d.Paths.PathItems.Get("/posts/{id}")
	if postsItem.Put == nil || postsItem.Put.Summary != "update" {
		t.Errorf("expected PUT /posts/{id} summary 'update'")
	}
}

func TestRouter_Path_WithMuxPrefix(t *testing.T) {
	base := http.NewServeMux()
	prefixed := httpx.Prefix(base, "/api")
	r := NewRouter("API", "0.0.0").WithMux(prefixed)

	r.Operation("GET /users", op.Summary("list"))

	d := r.GetDocument()
	if _, ok := d.Paths.PathItems.Get("/api/users"); !ok {
		t.Error("expected path /api/users with mux prefix")
	}
	if _, ok := d.Paths.PathItems.Get("/users"); ok {
		t.Error("expected path /users without prefix to not exist")
	}
}

func TestRouter_Path_NormalizesWildcard(t *testing.T) {
	r := NewRouter("API", "0.0.0")
	r.Operation("GET /files/{path...}", op.Summary("download"))

	if _, ok := r.GetDocument().Paths.PathItems.Get("/files/{path}"); !ok {
		t.Error("expected normalized path /files/{path}")
	}
}

func TestUseRouter_Found(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base, "Test", "1.0")

	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected UseRouter to return a non-nil Router")
	}
	if r.GetDocument().Info.Title != "Test" {
		t.Errorf("expected title 'Test', got %q", r.GetDocument().Info.Title)
	}
}

func TestUseRouter_NotFound(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected UseRouter to panic when no openapiMux is in the chain")
		}
	}()
	UseRouter(http.NewServeMux())
}

func TestUseRouter_ThroughPrefix(t *testing.T) {
	base := http.NewServeMux()
	withRouter := WithRouter(base, "Prefixed", "1.0")
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
	wrapped := WithRouter(base, "API", "0.0.0")

	r1 := UseRouter(wrapped)
	r2 := UseRouter(wrapped)

	if r1 == r2 {
		t.Fatal("expected UseRouter to return different instances")
	}
	if r1.GetDocument() != r2.GetDocument() {
		t.Error("expected UseRouter to share the same document across calls")
	}
}

func TestUseRouter_BindsOriginalMux(t *testing.T) {
	// UseRouter binds the returned Router to the mux passed in (not the inner one).
	// Operations registered via the Router should use that mux's prefix.
	base := http.NewServeMux()
	withRouter := WithRouter(base, "API", "1.0")
	prefixed := httpx.Prefix(withRouter, "/v1")

	r := UseRouter(prefixed)
	if r == nil {
		t.Fatal("expected non-nil router")
	}

	r.Operation("GET /users", op.Summary("list"))

	if _, ok := r.GetDocument().Paths.PathItems.Get("/v1/users"); !ok {
		t.Error("expected path /v1/users using the bound mux prefix")
	}
}
