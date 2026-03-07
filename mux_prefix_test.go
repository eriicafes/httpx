package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPrefix_HandleFunc(t *testing.T) {
	mux := Prefix(http.NewServeMux(), "/api")

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("healthy"))
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "healthy" {
		t.Errorf("expected body 'healthy', got %q", rec.Body.String())
	}
}

func TestPrefix_WithMethod(t *testing.T) {
	mux := Prefix(http.NewServeMux(), "/api")

	mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get users"))
	})
	mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create user"))
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "get users" {
		t.Errorf("GET: expected body 'get users', got %q", rec.Body.String())
	}

	req = httptest.NewRequest("POST", "/api/users", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "create user" {
		t.Errorf("POST: expected body 'create user', got %q", rec.Body.String())
	}
}

func TestPrefix_NestedPrefix(t *testing.T) {
	mux := Prefix(Prefix(http.NewServeMux(), "/api"), "/v1")

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("v1 users"))
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "v1 users" {
		t.Errorf("expected body 'v1 users', got %q", rec.Body.String())
	}
}

func TestGroup_HandleFunc(t *testing.T) {
	mux := http.NewServeMux()
	Group(mux, "/api", func(mux Mux) {
		mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("healthy"))
		})
	})

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "healthy" {
		t.Errorf("expected body 'healthy', got %q", rec.Body.String())
	}
}

func TestGroup_WithMethod(t *testing.T) {
	mux := http.NewServeMux()
	Group(mux, "/api", func(mux Mux) {
		mux.HandleFunc("GET /users", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("get users"))
		})
		mux.HandleFunc("POST /users", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("create user"))
		})
	})

	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("GET: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "get users" {
		t.Errorf("GET: expected body 'get users', got %q", rec.Body.String())
	}

	req = httptest.NewRequest("POST", "/api/users", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("POST: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "create user" {
		t.Errorf("POST: expected body 'create user', got %q", rec.Body.String())
	}
}

func TestGroup_NestedGroup(t *testing.T) {
	mux := http.NewServeMux()
	Group(mux, "/api", func(mux Mux) {
		Group(mux, "/v1", func(mux Mux) {
			mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
				w.Write([]byte("v1 users"))
			})
		})
	})

	req := httptest.NewRequest("GET", "/api/v1/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "v1 users" {
		t.Errorf("expected body 'v1 users', got %q", rec.Body.String())
	}
}

func TestMuxPrefix(t *testing.T) {
	mux := http.NewServeMux()
	tests := []struct {
		name     string
		mux      func() ServeMux
		expected string
	}{
		{
			name:     "no prefix",
			mux:      func() ServeMux { return mux },
			expected: "",
		},
		{
			name:     "single prefix",
			mux:      func() ServeMux { return Prefix(mux, "/api") },
			expected: "/api",
		},
		{
			name:     "nested prefix",
			mux:      func() ServeMux { return Prefix(Prefix(mux, "/api"), "/v1") },
			expected: "/api/v1",
		},
		{
			name:     "in-between prefix",
			mux:      func() ServeMux { return Prefix(Use(Prefix(mux, "/api")), "/v1") },
			expected: "/api/v1",
		},
		{
			name:     "deep prefix",
			mux:      func() ServeMux { return NormalizeTrailingSlash(Use(Prefix(mux, "/api"))) },
			expected: "/api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MuxPrefix(tt.mux())
			if result != tt.expected {
				t.Errorf("MuxPrefix() = %q, expected %q", result, tt.expected)
			}
		})
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

		// pattern with method only
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
		mux := http.NewServeMux()
		t.Run(tt.name, func(t *testing.T) {
			m := &prefixMux{mux, tt.prefix}
			result := m.prefixPattern(tt.pattern)
			if result != tt.expected {
				t.Errorf("prefixPattern(%q + %q) = %q, expected %q", tt.prefix, tt.pattern, result, tt.expected)
			}
		})
	}
}
