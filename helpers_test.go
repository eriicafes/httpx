package httpx

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSend(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]string{
		"message": "hello",
		"status":  "success",
	}

	if err := Send(rec, data); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["message"] != "hello" {
		t.Errorf("expected message 'hello', got %q", response["message"])
	}

	if response["status"] != "success" {
		t.Errorf("expected status 'success', got %q", response["status"])
	}
}

func TestSendStatus(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]string{
		"id":   "123",
		"name": "Test User",
	}

	if err := SendStatus(rec, http.StatusCreated, data); err != nil {
		t.Fatalf("SendStatus failed: %v", err)
	}

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}

	var response map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response["id"] != "123" {
		t.Errorf("expected id '123', got %q", response["id"])
	}

	if response["name"] != "Test User" {
		t.Errorf("expected name 'Test User', got %q", response["name"])
	}
}

func TestRead(t *testing.T) {
	type TestData struct {
		Name string `json:"name"`
	}

	tests := []struct {
		name          string
		body          string
		contentType   string
		expectError   bool
		errorContains string
		expectedName  string
	}{
		{
			name:         "valid JSON",
			body:         `{"name":"John"}`,
			contentType:  "application/json",
			expectedName: "John",
		},
		{
			name:          "empty body",
			body:          "",
			contentType:   "application/json",
			expectError:   true,
			errorContains: "invalid JSON",
		},
		{
			name:          "invalid JSON",
			body:          `{"name":"John",`,
			contentType:   "application/json",
			expectError:   true,
			errorContains: "invalid JSON",
		},
		{
			name:          "wrong content type",
			body:          `{"name":"John"}`,
			contentType:   "text/plain",
			expectError:   true,
			errorContains: "invalid content-type text/plain",
		},
		{
			name:         "missing content type",
			body:         `{"name":"John"}`,
			contentType:  "",
			expectedName: "John",
		},
		{
			name:         "content type with charset",
			body:         `{"name":"Jane"}`,
			contentType:  "application/json; charset=utf-8",
			expectedName: "Jane",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/test", strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			data, err := Read[TestData](req)

			if tt.expectError {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("error %q does not contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if data.Name != tt.expectedName {
					t.Errorf("expected name %s, got %s", tt.expectedName, data.Name)
				}
			}
		})
	}
}
