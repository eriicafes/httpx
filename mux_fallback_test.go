package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFallback_HandleError(t *testing.T) {
	baseMux := http.NewServeMux()
	errorHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("fallback: " + err.Error()))
	})

	mux := Fallback(baseMux, errorHandler)

	mux.Route("/error", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}

	if rec.Body.String() != "fallback: test error" {
		t.Errorf("expected body 'fallback: test error', got %q", rec.Body.String())
	}
}

func TestFallback_HandleSuccess(t *testing.T) {
	baseMux := http.NewServeMux()
	errorHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("should not reach here"))
	})

	mux := Fallback(baseMux, errorHandler)

	mux.Route("/success", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("success"))
		return nil
	})

	req := httptest.NewRequest("GET", "/success", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %q", rec.Body.String())
	}
}

func TestFallback_Handle(t *testing.T) {
	baseMux := http.NewServeMux()
	errorHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
	})

	mux := Fallback(baseMux, errorHandler)

	mux.Handle("/test", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("test"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "test" {
		t.Errorf("expected body 'test', got %q", rec.Body.String())
	}
}

func TestFallback_HandleFunc(t *testing.T) {
	baseMux := http.NewServeMux()
	errorHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
	})

	mux := Fallback(baseMux, errorHandler)

	mux.HandleFunc("/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	})

	req := httptest.NewRequest("GET", "/hello", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "hello" {
		t.Errorf("expected body 'hello', got %q", rec.Body.String())
	}
}

func TestFallback_CustomJSONErrorHandler(t *testing.T) {
	baseMux := http.NewServeMux()
	errorHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":  err.Error(),
			"custom": "field",
		})
	})

	mux := Fallback(baseMux, errorHandler)

	mux.Route("/json-error", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("json error")
	})

	req := httptest.NewRequest("GET", "/json-error", nil)
	rec := httptest.NewRecorder()

	baseMux.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] != "json error" {
		t.Errorf("expected error 'json error', got %q", response["error"])
	}

	if response["custom"] != "field" {
		t.Errorf("expected custom field 'field', got %q", response["custom"])
	}
}
