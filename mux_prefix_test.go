package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrefix_HandleFunc(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := Prefix(baseMux, "/api")

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthy"))
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "healthy" {
		t.Errorf("expected body 'healthy', got %q", rec.Body.String())
	}
}

func TestPrefix_Handle(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := Prefix(baseMux, "/api/v1")

	mux.Handle("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users handler"))
	}))

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "users handler" {
		t.Errorf("expected body 'users handler', got %q", rec.Body.String())
	}
}

func TestPrefix_Route(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := Prefix(baseMux, "/api")

	mux.Route("/data", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("data"))
		return nil
	})

	req := httptest.NewRequest("GET", "/api/data", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "data" {
		t.Errorf("expected body 'data', got %q", rec.Body.String())
	}
}

func TestPrefix_WithMethod(t *testing.T) {
	baseMux := http.NewServeMux()
	mux := Prefix(baseMux, "/api/v2")

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get users"))
	})

	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create user"))
	})

	// Test GET
	req := httptest.NewRequest("GET", "/api/v2/users", nil)
	rec := httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "get users" {
		t.Errorf("GET: expected body 'get users', got %q", rec.Body.String())
	}

	// Test POST
	req = httptest.NewRequest("POST", "/api/v2/users", nil)
	rec = httptest.NewRecorder()
	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "create user" {
		t.Errorf("POST: expected body 'create user', got %q", rec.Body.String())
	}
}

func TestPrefix_NestedPrefix(t *testing.T) {
	baseMux := http.NewServeMux()
	apiMux := Prefix(baseMux, "/api")
	v1Mux := Prefix(apiMux, "/v1")

	v1Mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1 users"))
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "v1 users" {
		t.Errorf("expected body 'v1 users', got %q", rec.Body.String())
	}
}

func TestPrefix_PrefixPattern(t *testing.T) {

	tests := []struct {
		name     string
		prefix   string
		pattern  string
		expected string
	}{
		{
			name:     "simple pattern",
			prefix:   "/api",
			pattern:  "/users",
			expected: "/api/users",
		},
		{
			name:     "pattern with method",
			prefix:   "/api",
			pattern:  "GET /users",
			expected: "GET /api/users",
		},
		{
			name:     "pattern with POST method",
			prefix:   "/api",
			pattern:  "POST /items/{id}",
			expected: "POST /api/items/{id}",
		},
		{
			name:     "pattern with wildcard",
			prefix:   "/api",
			pattern:  "/files/{path...}",
			expected: "/api/files/{path...}",
		},
		{
			name:     "pattern with exact match",
			prefix:   "/api",
			pattern:  "/users/{$}",
			expected: "/api/users/{$}",
		},

		// trailing slash patterns
		{
			name:     "root pattern",
			prefix:   "/api",
			pattern:  "/",
			expected: "/api/",
		},
		{
			name:     "root prefix",
			prefix:   "/",
			pattern:  "api",
			expected: "/api",
		},
		{
			name:     "root prefix with empty pattern",
			prefix:   "/",
			pattern:  "",
			expected: "/",
		},
		{
			name:     "prefix with trailing slash",
			prefix:   "/api/",
			pattern:  "users",
			expected: "/api/users",
		},
		{
			name:     "pattern with trailing slash",
			prefix:   "/api",
			pattern:  "/users/",
			expected: "/api/users/",
		},
		{
			name:     "both have trailing slash",
			prefix:   "/api/",
			pattern:  "users/",
			expected: "/api/users/",
		},

		// pattern with method
		{
			name:     "root prefix with method",
			prefix:   "/",
			pattern:  "GET ",
			expected: "GET /",
		},
		{
			name:     "prefix with method",
			prefix:   "/api",
			pattern:  "GET ",
			expected: "GET /api",
		},
		{
			name:     "prefix with root pattern and method",
			prefix:   "/api",
			pattern:  "GET /",
			expected: "GET /api/",
		},
		{
			name:     "prefix with pattern and method",
			prefix:   "/api",
			pattern:  "GET /users",
			expected: "GET /api/users",
		},
	}

	for _, tt := range tests {
		baseMux := http.NewServeMux()
		t.Run(tt.name, func(t *testing.T) {
			prefixMux := &prefixMux{baseMux, tt.prefix}

			result := prefixMux.prefixPattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("prefixPattern(%q + %q) = %q, expected %q", tt.prefix, tt.pattern, result, tt.expected)
			}
		})
	}
}
