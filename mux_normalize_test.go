package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeTrailingSlash_BothPaths(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users"))
	})

	// Test without trailing slash
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("without slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "users" {
		t.Errorf("without slash: expected body 'users', got %q", rec.Body.String())
	}

	// Test with trailing slash
	req = httptest.NewRequest("GET", "/users/", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("with slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "users" {
		t.Errorf("with slash: expected body 'users', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_PatternWithSlash(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	// Pattern already ends with slash, should not duplicate
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api"))
	})

	req := httptest.NewRequest("GET", "/api/", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "api" {
		t.Errorf("expected body 'api', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_PatternWithExactMatch(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	// Pattern already ends with {$}, should not duplicate
	mux.HandleFunc("/exact/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("exact"))
	})

	req := httptest.NewRequest("GET", "/exact/", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "exact" {
		t.Errorf("expected body 'exact', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_Handle(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	mux.Handle("/products", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("products"))
	}))

	// Test without trailing slash
	req := httptest.NewRequest("GET", "/products", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "products" {
		t.Errorf("without slash: expected body 'products', got %q", rec.Body.String())
	}

	// Test with trailing slash
	req = httptest.NewRequest("GET", "/products/", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "products" {
		t.Errorf("with slash: expected body 'products', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_Route(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	mux.Route("/items", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("items"))
		return nil
	})

	// Test without trailing slash
	req := httptest.NewRequest("GET", "/items", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "items" {
		t.Errorf("without slash: expected body 'items', got %q", rec.Body.String())
	}

	// Test with trailing slash
	req = httptest.NewRequest("GET", "/items/", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "items" {
		t.Errorf("with slash: expected body 'items', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_WithMethod(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := NormalizeTrailingSlash(baseMux)

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get users"))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create user"))
	})

	// Test GET without trailing slash
	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "get users" {
		t.Errorf("GET without slash: expected body 'get users', got %q", rec.Body.String())
	}

	// Test GET with trailing slash
	req = httptest.NewRequest("GET", "/api/users/", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "get users" {
		t.Errorf("GET with slash: expected body 'get users', got %q", rec.Body.String())
	}

	// Test POST without trailing slash
	req = httptest.NewRequest("POST", "/api/users", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "create user" {
		t.Errorf("POST without slash: expected body 'create user', got %q", rec.Body.String())
	}

	// Test POST with trailing slash
	req = httptest.NewRequest("POST", "/api/users/", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Body.String() != "create user" {
		t.Errorf("POST with slash: expected body 'create user', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_Patterns(t *testing.T) {
	baseMux := http.NewServeMux()
	normalizeMux := &normalizeTrailingSlashMux{baseMux}

	tests := []struct {
		name     string
		pattern  string
		pattern1 string
		pattern2 string
	}{
		{
			name:     "simple pattern",
			pattern:  "/users",
			pattern1: "/users",
			pattern2: "/users/{$}",
		},
		{
			name:     "pattern with trailing slash",
			pattern:  "/api/",
			pattern1: "/api/",
			pattern2: "",
		},
		{
			name:     "pattern with exact match",
			pattern:  "/exact/{$}",
			pattern1: "/exact/{$}",
			pattern2: "",
		},
		{
			name:     "pattern with method",
			pattern:  "GET /users",
			pattern1: "GET /users",
			pattern2: "GET /users/{$}",
		},
		{
			name:     "pattern with wildcard",
			pattern:  "/files/{path...}",
			pattern1: "/files/{path...}",
			pattern2: "/files/{path...}/{$}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p1, p2 := normalizeMux.patterns(tt.pattern)
			if p1 != tt.pattern1 {
				t.Errorf("pattern1: expected %q, got %q", tt.pattern1, p1)
			}
			if p2 != tt.pattern2 {
				t.Errorf("pattern2: expected %q, got %q", tt.pattern2, p2)
			}
		})
	}
}
