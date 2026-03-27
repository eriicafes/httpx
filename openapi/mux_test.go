package openapi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/doc"
)

func TestWithRouter_HandleFunc(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base, "API", "0.0.0")

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
	wrapped := WithRouter(base, "API", "0.0.0")

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
	wrapped := WithRouter(base, "API", "0.0.0")

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
	wrapped := WithRouter(base, "API", "0.0.0")
	wrapper2 := httpx.Use(wrapped)

	// UseRouter traverses the chain — verify the inner mux is accessible.
	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected UseRouter to extract Router from WithRouter mux")
	}
	r = UseRouter(wrapper2)
	if r == nil {
		t.Fatal("expected UseRouter to extract Router from WithRouter mux")
	}
}

func TestWithRouter_OpenapiDocOptions(t *testing.T) {
	base := http.NewServeMux()
	wrapped := WithRouter(base, "API", "0.0.0", doc.Version("3.1.1"))

	r := UseRouter(wrapped)
	if r == nil {
		t.Fatal("expected non-nil router")
	}
	d := r.GetDocument()
	if d.Info.Title != "API" {
		t.Errorf("expected title 'API', got %q", d.Info.Title)
	}
	if d.Info.Version != "0.0.0" {
		t.Errorf("expected version '0.0.0', got %q", d.Info.Version)
	}
	if d.Version != "3.1.1" {
		t.Errorf("expected openapi version '3.1.1', got %q", d.Version)
	}
}
