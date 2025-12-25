package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/eriicafes/httpx/httperrors"
)

var handlerTests = []struct {
	name           string
	handler        func(http.ResponseWriter, *http.Request) error
	expectedStatus int
	expectedError  string
	expectedBody   string
	expectedCode   string
}{
	{
		name: "success",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("success"))
			return nil
		},
		expectedStatus: http.StatusOK,
		expectedBody:   "success",
	},
	{
		name: "with error",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			return errors.New("test error")
		},
		expectedStatus: http.StatusBadRequest,
		expectedError:  "test error",
	},
	{
		name: "with HTTP error",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			return httperrors.New("forbidden", http.StatusForbidden)
		},
		expectedStatus: http.StatusForbidden,
		expectedError:  "forbidden",
	},
	{
		name: "with HTTP error with code",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			return httperrors.New("validation failed", http.StatusUnprocessableEntity, httperrors.Code("VALIDATION_ERROR"))
		},
		expectedStatus: http.StatusUnprocessableEntity,
		expectedError:  "validation failed",
		expectedCode:   "VALIDATION_ERROR",
	},
	{
		name: "with internal error",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			dbErr := errors.New("database connection failed")
			return InternalError(dbErr, "Failed to fetch user")
		},
		expectedStatus: http.StatusInternalServerError,
		expectedError:  "Failed to fetch user",
	},
	{
		name: "error handler clears content-length header",
		handler: func(w http.ResponseWriter, r *http.Request) error {
			w.Header().Set("Content-Length", "12345")
			w.Header().Set("X-Custom-Header", "should-remain")
			return errors.New("test error")
		},
		expectedStatus: http.StatusBadRequest,
		expectedError:  "test error",
	},
}

func TestHandlerFunc(t *testing.T) {
	tests := handlerTests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := HandlerFunc(tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			handler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" {
				if rec.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rec.Body.String())
				}
				return
			}

			// Verify error handler sets correct headers
			if rec.Header().Get("Content-Length") != "" {
				t.Errorf("expected Content-Length to be cleared, got %q", rec.Header().Get("Content-Length"))
			}

			if rec.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
			}

			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("expected X-Content-Type-Options 'nosniff', got %q", rec.Header().Get("X-Content-Type-Options"))
			}

			if tt.expectedError != "" {
				var response map[string]any
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response["error"] != tt.expectedError {
					t.Errorf("expected error message %q, got %q", tt.expectedError, response["error"])
				}

				if tt.expectedCode != "" {
					if response["code"] != tt.expectedCode {
						t.Errorf("expected code %q, got %q", tt.expectedCode, response["code"])
					}
				}
			}
		})
	}
}

func TestMux_Route(t *testing.T) {
	tests := handlerTests

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := New()
			mux.Route("GET /test", tt.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			rec := httptest.NewRecorder()

			mux.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if tt.expectedBody != "" {
				if rec.Body.String() != tt.expectedBody {
					t.Errorf("expected body %q, got %q", tt.expectedBody, rec.Body.String())
				}
				return
			}

			// Verify error handler sets correct headers
			if rec.Header().Get("Content-Length") != "" {
				t.Errorf("expected Content-Length to be cleared, got %q", rec.Header().Get("Content-Length"))
			}

			if rec.Header().Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
			}

			if rec.Header().Get("X-Content-Type-Options") != "nosniff" {
				t.Errorf("expected X-Content-Type-Options 'nosniff', got %q", rec.Header().Get("X-Content-Type-Options"))
			}

			if tt.expectedError != "" {
				var response map[string]any
				if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if response["error"] != tt.expectedError {
					t.Errorf("expected error message %q, got %q", tt.expectedError, response["error"])
				}

				if tt.expectedCode != "" {
					if response["code"] != tt.expectedCode {
						t.Errorf("expected code %q, got %q", tt.expectedCode, response["code"])
					}
				}
			}
		})
	}
}

func TestInternalError(t *testing.T) {
	tests := []struct {
		name            string
		underlyingError error
		message         string
		expectedMessage string
	}{
		{
			name:            "empty message uses default",
			underlyingError: errors.New("underlying error"),
			message:         "",
			expectedMessage: "Something went wrong",
		},
		{
			name:            "custom message",
			underlyingError: errors.New("database error"),
			message:         "Failed to save",
			expectedMessage: "Failed to save",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InternalError(tt.underlyingError, tt.message)

			if err.Error() != tt.expectedMessage {
				t.Errorf("expected message %q, got %q", tt.expectedMessage, err.Error())
			}

			unwrapped := errors.Unwrap(err)
			if unwrapped != tt.underlyingError {
				t.Errorf("expected unwrapped error to be %v, got %v", tt.underlyingError, unwrapped)
			}
		})
	}
}

func TestMuxErrorHandler(t *testing.T) {
	baseMux := http.NewServeMux()
	customHandler := ErrorHandlerFunc(func(w http.ResponseWriter, r *http.Request, err error) {
		w.WriteHeader(http.StatusTeapot)
		w.Write([]byte("custom error"))
	})
	mux := Fallback(baseMux, customHandler)

	handler := MuxErrorHandler(mux)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	handler.HandleError(rec, req, errors.New("test error"))

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}

	if rec.Body.String() != "custom error" {
		t.Errorf("expected body 'custom error', got %q", rec.Body.String())
	}
}

func TestApplyMuxErrorHandler(t *testing.T) {
	baseMux := http.NewServeMux()

	handler := func(w http.ResponseWriter, r *http.Request) error {
		return errors.New("handler error")
	}

	wrappedHandler := Handler(baseMux, handler)

	req := httptest.NewRequest("GET", "/test", nil)
	rec := httptest.NewRecorder()

	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	// Verify it used the mux's error handler
	var response map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["error"] != "handler error" {
		t.Errorf("expected error message 'handler error', got %q", response["error"])
	}
}
