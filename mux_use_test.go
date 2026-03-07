package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUse_SingleMiddleware(t *testing.T) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	mux := Use(http.NewServeMux(), middleware)

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Header().Get("X-Middleware") != "applied" {
		t.Errorf("expected X-Middleware header 'applied', got %q", rec.Header().Get("X-Middleware"))
	}

	if rec.Body.String() != "test" {
		t.Errorf("expected body 'test', got %q", rec.Body.String())
	}
}

func TestUse_MultipleMiddlewares(t *testing.T) {
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-1", "first")
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-2", "second")
			next.ServeHTTP(w, r)
		})
	}

	middleware3 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Middleware-3", "third")
			next.ServeHTTP(w, r)
		})
	}

	mux := Use(http.NewServeMux(), middleware1, middleware2, middleware3)

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Header().Get("X-Middleware-1") != "first" {
		t.Errorf("expected X-Middleware-1 header 'first', got %q", rec.Header().Get("X-Middleware-1"))
	}

	if rec.Header().Get("X-Middleware-2") != "second" {
		t.Errorf("expected X-Middleware-2 header 'second', got %q", rec.Header().Get("X-Middleware-2"))
	}

	if rec.Header().Get("X-Middleware-3") != "third" {
		t.Errorf("expected X-Middleware-3 header 'third', got %q", rec.Header().Get("X-Middleware-3"))
	}

	if rec.Body.String() != "test" {
		t.Errorf("expected body 'test', got %q", rec.Body.String())
	}
}

func TestUse_MiddlewareOrder(t *testing.T) {
	var executionOrder []string

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m1-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m1-after")
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			executionOrder = append(executionOrder, "m2-before")
			next.ServeHTTP(w, r)
			executionOrder = append(executionOrder, "m2-after")
		})
	}

	mux := Use(http.NewServeMux(), middleware1, middleware2)

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		executionOrder = append(executionOrder, "handler")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	expected := []string{"m1-before", "m2-before", "handler", "m2-after", "m1-after"}
	if len(executionOrder) != len(expected) {
		t.Fatalf("expected %d executions, got %d", len(expected), len(executionOrder))
	}

	for i, exp := range expected {
		if executionOrder[i] != exp {
			t.Errorf("execution order[%d]: expected %q, got %q", i, exp, executionOrder[i])
		}
	}
}

func TestUse_NoMiddlewares(t *testing.T) {
	mux := Use(http.NewServeMux())

	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("no middleware"))
	})

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Body.String() != "no middleware" {
		t.Errorf("expected body 'no middleware', got %q", rec.Body.String())
	}
}

func TestUse_MiddlewareShortCircuit(t *testing.T) {
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Short-circuit: don't call next handler
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("unauthorized"))
		})
	}

	mux := Use(http.NewServeMux(), middleware)

	mux.HandleFunc("/protected", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("should not reach here"))
	})

	req := httptest.NewRequest("GET", "/protected", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	if rec.Body.String() != "unauthorized" {
		t.Errorf("expected body 'unauthorized', got %q", rec.Body.String())
	}
}
