package openapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eriicafes/httpx/openapi/doc"
	"github.com/eriicafes/httpx/openapi/op"
)

func TestWithRouter_ImplementsServeMux(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base)

	// WithRouter should implement httpx.ServeMux (Handle, HandleFunc, ServeHTTP).
	wrapped.Handle("GET /test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestWithRouter_HandleFunc(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base)

	wrapped.HandleFunc("GET /ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	req := httptest.NewRequest("GET", "/ping", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "pong" {
		t.Errorf("expected body 'pong', got %q", rec.Body.String())
	}
}

func TestWithRouter_Route(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base)

	wrapped.Route("GET /hello", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("hello"))
		return nil
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "hello" {
		t.Errorf("expected body 'hello', got %q", rec.Body.String())
	}
}

func TestWithRouter_DelegatesToInnerMux(t *testing.T) {
	base := http.NewServeMux()
	base.HandleFunc("GET /existing", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("from base"))
	})
	wrapped := WithRouter(base)

	req := httptest.NewRequest("GET", "/existing", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "from base" {
		t.Errorf("expected body 'from base', got %q", rec.Body.String())
	}
}

func TestWithRouter_Mux(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base)

	// UseRouter traverses the chain — verify the inner mux is accessible.
	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected UseRouter to extract Router from WithRouter mux")
	}
}

func TestWithRouter_OpenapiDocOptions(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base,
		doc.Info("Test Service", "1.2.3"),
		doc.OpenAPIVersion("3.1.0"),
	)

	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	d := r.GetDocument()
	if d.Info.Title != "Test Service" {
		t.Errorf("expected title 'Test Service', got %q", d.Info.Title)
	}
	if d.Info.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", d.Info.Version)
	}
}

func TestWithRouter_FullFlow(t *testing.T) {
	// Full integration: WithRouter wraps a mux, UseRouter extracts a Router,
	// operations are registered, and routes are served.
	base := http.NewServeMux()
	wrapped := WithRouter(base, doc.Info("Full Flow", "1.0"))

	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected non-nil router")
	}

	r.Route("GET /users", op.Summary("list users"), func(w http.ResponseWriter, req *http.Request) error {
		w.Write([]byte("users"))
		return nil
	})

	// The route is served via the wrapped mux.
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", rec.Code)
	}
	if rec.Body.String() != "users" {
		t.Errorf("expected body 'users', got %q", rec.Body.String())
	}

	// The operation is also recorded in the document.
	item, ok := r.GetDocument().Paths.PathItems.Get("/users")
	if !ok {
		t.Fatal("expected /users path item in document")
	}
	if item.Get == nil || item.Get.Summary != "list users" {
		t.Error("expected GET /users operation with summary 'list users'")
	}
}
