package httpx

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizeTrailingSlash_BothPaths(t *testing.T) {
	mux := NormalizeTrailingSlash(http.NewServeMux())

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("users"))
	})

	// Test without trailing slash
	req := httptest.NewRequest("GET", "/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("without slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "users" {
		t.Errorf("without slash: expected body 'users', got %q", rec.Body.String())
	}

	// Test with trailing slash
	req = httptest.NewRequest("GET", "/users/", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("with slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "users" {
		t.Errorf("with slash: expected body 'users', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_PatternWithSlash(t *testing.T) {
	mux := NormalizeTrailingSlash(http.NewServeMux())

	// Pattern already ends with slash, should not duplicate
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("api"))
	})

	// Test with trailing slash
	req := httptest.NewRequest("GET", "/api/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("with slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "api" {
		t.Errorf("with slash: expected body 'api', got %q", rec.Body.String())
	}

	// Test without trailing slash — subtree pattern redirects to trailing slash
	req = httptest.NewRequest("GET", "/api", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("without slash: expected status %d, got %d", http.StatusMovedPermanently, rec.Code)
	}

	// Test nested path — subtree pattern matches any path under /api/
	req = httptest.NewRequest("GET", "/api/nested", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("nested path: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "api" {
		t.Errorf("nested path: expected body 'api', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_PatternWithExactMatch(t *testing.T) {
	mux := NormalizeTrailingSlash(http.NewServeMux())

	// Pattern already ends with {$}, should not duplicate
	mux.HandleFunc("/exact/{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("exact"))
	})

	// Test with trailing slash — {$} anchors to this exact path
	req := httptest.NewRequest("GET", "/exact/", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("with slash: expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "exact" {
		t.Errorf("with slash: expected body 'exact', got %q", rec.Body.String())
	}

	// Test without trailing slash — mux redirects to trailing slash (same as PatternWithSlash)
	req = httptest.NewRequest("GET", "/exact", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMovedPermanently {
		t.Errorf("without slash: expected status %d, got %d", http.StatusMovedPermanently, rec.Code)
	}

	// Test nested path — should not match since {$} prevents subtree matching
	req = httptest.NewRequest("GET", "/exact/nested", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("nested path: expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestNormalizeTrailingSlash_WithMethod(t *testing.T) {
	mux := NormalizeTrailingSlash(http.NewServeMux())

	mux.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("get users"))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("create user"))
	})

	// Test GET without trailing slash
	req := httptest.NewRequest("GET", "/api/users", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Body.String() != "get users" {
		t.Errorf("GET without slash: expected body 'get users', got %q", rec.Body.String())
	}

	// Test GET with trailing slash
	req = httptest.NewRequest("GET", "/api/users/", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Body.String() != "get users" {
		t.Errorf("GET with slash: expected body 'get users', got %q", rec.Body.String())
	}

	// Test POST without trailing slash
	req = httptest.NewRequest("POST", "/api/users", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Body.String() != "create user" {
		t.Errorf("POST without slash: expected body 'create user', got %q", rec.Body.String())
	}

	// Test POST with trailing slash
	req = httptest.NewRequest("POST", "/api/users/", nil)
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Body.String() != "create user" {
		t.Errorf("POST with slash: expected body 'create user', got %q", rec.Body.String())
	}
}

func TestNormalizeTrailingSlash_Patterns(t *testing.T) {
	normalizeMux := &normalizeTrailingSlashMux{http.NewServeMux()}

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
