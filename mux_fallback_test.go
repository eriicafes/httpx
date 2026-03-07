package httpx

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFallback_HandleError(t *testing.T) {
	mux := Fallback(http.NewServeMux(), func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("fallback: " + err.Error()))
	})

	mux.Route("/error", func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("test error")
	})

	req := httptest.NewRequest("GET", "/error", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Errorf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}

	if rec.Body.String() != "fallback: test error" {
		t.Errorf("expected body 'fallback: test error', got %q", rec.Body.String())
	}
}

func TestFallback_HandleSuccess(t *testing.T) {
	mux := Fallback(http.NewServeMux(), func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("should not reach here"))
	})

	mux.Route("/success", func(w http.ResponseWriter, r *http.Request) error {
		w.Write([]byte("success"))
		return nil
	})

	req := httptest.NewRequest("GET", "/success", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if rec.Body.String() != "success" {
		t.Errorf("expected body 'success', got %q", rec.Body.String())
	}
}
