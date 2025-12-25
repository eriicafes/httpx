package httperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	t.Run("without data", func(t *testing.T) {
		err := New("User not found", http.StatusNotFound)

		if err.Message() != "User not found" {
			t.Errorf("expected message 'User not found', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, err.StatusCode())
		}
		if err.ErrorData() != nil {
			t.Errorf("expected nil data, got %v", err.ErrorData())
		}
		if err.Error() != "User not found" {
			t.Errorf("expected error string 'User not found', got '%s'", err.Error())
		}
	})

	t.Run("with Details", func(t *testing.T) {
		details := Details{
			"email":    "must be a valid email address",
			"password": "must be at least 8 characters",
		}
		err := New("Invalid input", http.StatusBadRequest, details)

		if err.Message() != "Invalid input" {
			t.Errorf("expected message 'Invalid input', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
		}
		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if len(detailsMap) != 2 {
			t.Errorf("expected 2 details, got %d", len(detailsMap))
		}
		if detailsMap["email"] != "must be a valid email address" {
			t.Errorf("unexpected email detail: %s", detailsMap["email"])
		}
		if detailsMap["password"] != "must be at least 8 characters" {
			t.Errorf("unexpected password detail: %s", detailsMap["password"])
		}
	})

	t.Run("with Code", func(t *testing.T) {
		code := Code("NOT_FOUND")
		err := New("User not found", http.StatusNotFound, code)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code %s, got %v", code, data["code"])
		}
	})

	t.Run("with Fields", func(t *testing.T) {
		fields := Fields{
			"request_id": "req-123",
			"user_id":    456,
		}
		err := New("Internal error", http.StatusInternalServerError, fields)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["request_id"] != "req-123" {
			t.Errorf("unexpected request_id: %v", data["request_id"])
		}
		if data["user_id"] != 456 {
			t.Errorf("unexpected user_id: %v", data["user_id"])
		}
	})

	t.Run("with mixed metadata", func(t *testing.T) {
		code := Code("VALIDATION_ERROR")
		details := Details{"email": "invalid"}
		fields := Fields{"request_id": "req-789"}

		err := New("Validation failed", http.StatusBadRequest, code, details, fields)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code in data")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Errorf("expected details in data")
		}
		if detailsMap["email"] != "invalid" {
			t.Errorf("expected email detail")
		}
		if data["request_id"] != "req-789" {
			t.Errorf("expected request_id in data")
		}
	})
}

func TestReport(t *testing.T) {
	t.Run("without data", func(t *testing.T) {
		baseErr := errors.New("database connection failed")
		err := Report(baseErr, http.StatusInternalServerError)

		if err.Message() != "database connection failed" {
			t.Errorf("expected message 'database connection failed', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, err.StatusCode())
		}
		if err.Error() != "database connection failed" {
			t.Errorf("expected error string 'database connection failed', got '%s'", err.Error())
		}
		if !errors.Is(err, baseErr) {
			t.Error("expected error to unwrap to base error")
		}
	})

	t.Run("with metadata", func(t *testing.T) {
		baseErr := errors.New("validation failed")
		code := Code("VALIDATION_ERROR")
		details := Details{"username": "required field"}

		err := Report(baseErr, http.StatusBadRequest, code, details)

		if err.Message() != "validation failed" {
			t.Errorf("expected message 'validation failed', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
		}
		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code in data")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["username"] != "required field" {
			t.Errorf("unexpected username detail: %s", detailsMap["username"])
		}
		if !errors.Is(err, baseErr) {
			t.Error("expected error to unwrap to base error")
		}
	})
}

func TestWrap(t *testing.T) {
	t.Run("without data", func(t *testing.T) {
		baseErr := errors.New("connection timeout")
		err := Wrap(baseErr, "Payment processing failed", http.StatusBadGateway)

		if err.Message() != "Payment processing failed" {
			t.Errorf("expected message 'Payment processing failed', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadGateway {
			t.Errorf("expected status code %d, got %d", http.StatusBadGateway, err.StatusCode())
		}
		if err.Error() != "Payment processing failed: connection timeout" {
			t.Errorf("unexpected error string: %s", err.Error())
		}
		if !errors.Is(err, baseErr) {
			t.Error("expected error to unwrap to base error")
		}
	})

	t.Run("with metadata", func(t *testing.T) {
		baseErr := errors.New("validation error")
		code := Code("VALIDATION_ERROR")
		details := Details{"email": "already in use"}

		err := Wrap(baseErr, "Validation failed", http.StatusBadRequest, code, details)

		if err.Message() != "Validation failed" {
			t.Errorf("expected message 'Validation failed', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
		}
		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code in data")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["email"] != "already in use" {
			t.Errorf("unexpected email detail: %s", detailsMap["email"])
		}
		if err.Error() != "Validation failed: validation error" {
			t.Errorf("unexpected error string: %s", err.Error())
		}
	})

	t.Run("inherit status code from HTTPError", func(t *testing.T) {
		baseErr := New("Original error", http.StatusNotFound)
		err := Wrap(baseErr, "Wrapped error", 0)

		if err.StatusCode() != http.StatusNotFound {
			t.Errorf("expected status code %d (inherited), got %d", http.StatusNotFound, err.StatusCode())
		}
		if err.Message() != "Wrapped error" {
			t.Errorf("expected message 'Wrapped error', got '%s'", err.Message())
		}
	})

	t.Run("inherit message from HTTPError", func(t *testing.T) {
		baseErr := New("Original message", http.StatusNotFound)
		err := Wrap(baseErr, "", http.StatusBadRequest)

		if err.Message() != "Original message" {
			t.Errorf("expected message 'Original message' (inherited), got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
		}
	})

	t.Run("inherit data from HTTPError", func(t *testing.T) {
		baseDetails := Details{"field": "error"}
		baseErr := New("Original", http.StatusBadRequest, baseDetails)
		err := Wrap(baseErr, "Wrapped", http.StatusBadRequest)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected inherited error data, got nil")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["field"] != "error" {
			t.Errorf("expected inherited details, got %v", detailsMap)
		}
	})

	t.Run("override all from HTTPError", func(t *testing.T) {
		baseErr := New("Original", http.StatusNotFound, Details{"old": "value"})
		newDetails := Details{"new": "value"}
		err := Wrap(baseErr, "New message", http.StatusBadRequest, newDetails)

		if err.Message() != "New message" {
			t.Errorf("expected message 'New message', got '%s'", err.Message())
		}
		if err.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
		}
		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["new"] != "value" {
			t.Errorf("expected new details, got %v", detailsMap)
		}
	})
}

func TestUnwrap(t *testing.T) {
	t.Run("direct HTTPError", func(t *testing.T) {
		err := New("Test error", http.StatusBadRequest)
		httpErr, ok := Unwrap(err)

		if !ok {
			t.Error("expected Unwrap to return true")
		}
		if httpErr.Message() != "Test error" {
			t.Errorf("expected message 'Test error', got '%s'", httpErr.Message())
		}
	})

	t.Run("wrapped HTTPError", func(t *testing.T) {
		baseErr := New("Base error", http.StatusNotFound)
		wrappedErr := Wrap(baseErr, "Wrapped", http.StatusInternalServerError)
		httpErr, ok := Unwrap(wrappedErr)

		if !ok {
			t.Error("expected Unwrap to return true")
		}
		if httpErr.StatusCode() != http.StatusInternalServerError {
			t.Errorf("expected status code %d, got %d", http.StatusInternalServerError, httpErr.StatusCode())
		}
	})

	t.Run("non-HTTPError", func(t *testing.T) {
		err := errors.New("standard error")
		_, ok := Unwrap(err)

		if ok {
			t.Error("expected Unwrap to return false for standard error")
		}
	})

	t.Run("HTTPError wrapped by non-HTTPError", func(t *testing.T) {
		baseHTTPErr := New("Base HTTP error", http.StatusNotFound)
		intermediaryErr := fmt.Errorf("intermediary error: %w", baseHTTPErr)

		httpErr, ok := Unwrap(intermediaryErr)

		if !ok {
			t.Error("expected Unwrap to find HTTPError through intermediary error")
		}
		if httpErr.StatusCode() != http.StatusNotFound {
			t.Errorf("expected status code %d, got %d", http.StatusNotFound, httpErr.StatusCode())
		}
		if httpErr.Message() != "Base HTTP error" {
			t.Errorf("expected message 'Base HTTP error', got '%s'", httpErr.Message())
		}
	})

	t.Run("multiple non-HTTPError wrappers", func(t *testing.T) {
		baseHTTPErr := New("HTTP error", http.StatusBadRequest, Details{"field": "error"})
		err1 := fmt.Errorf("wrapper 1: %w", baseHTTPErr)
		err2 := fmt.Errorf("wrapper 2: %w", err1)
		err3 := fmt.Errorf("wrapper 3: %w", err2)

		httpErr, ok := Unwrap(err3)

		if !ok {
			t.Error("expected Unwrap to find HTTPError through multiple wrappers")
		}
		if httpErr.StatusCode() != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, httpErr.StatusCode())
		}
		if httpErr.Message() != "HTTP error" {
			t.Errorf("expected message 'HTTP error', got '%s'", httpErr.Message())
		}
		data := httpErr.ErrorData()
		if data == nil {
			t.Fatal("expected error data to be preserved, got nil")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["field"] != "error" {
			t.Errorf("expected details to be preserved, got %v", detailsMap)
		}
	})
}

func TestHTTPError(t *testing.T) {
	t.Run("with all values", func(t *testing.T) {
		details := Details{"field": "error"}
		err := New("Test message", http.StatusBadRequest, details)
		message, statusCode, data := err.HTTPError()

		if message != "Test message" {
			t.Errorf("expected message 'Test message', got '%s'", message)
		}
		if statusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, statusCode)
		}
		if data["error"] != "Test message" {
			t.Errorf("expected error field in data, got %v", data)
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["field"] != "error" {
			t.Errorf("unexpected details: %v", detailsMap)
		}
	})

	t.Run("with empty message", func(t *testing.T) {
		err := New("", http.StatusBadRequest)
		message, statusCode, _ := err.HTTPError()

		if message != "Something went wrong" {
			t.Errorf("expected default message 'Something went wrong', got '%s'", message)
		}
		if statusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, statusCode)
		}
	})

	t.Run("with zero status code", func(t *testing.T) {
		err := New("Test", 0)
		message, statusCode, _ := err.HTTPError()

		if message != "Test" {
			t.Errorf("expected message 'Test', got '%s'", message)
		}
		if statusCode != http.StatusInternalServerError {
			t.Errorf("expected default status code %d, got %d", http.StatusInternalServerError, statusCode)
		}
	})

	t.Run("with nil data adds error field", func(t *testing.T) {
		err := New("Test error", http.StatusBadRequest)
		message, statusCode, data := err.HTTPError()

		if message != "Test error" {
			t.Errorf("expected message 'Test error', got '%s'", message)
		}
		if statusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, statusCode)
		}
		if data == nil {
			t.Fatal("expected data map to be created, got nil")
		}
		if data["error"] != "Test error" {
			t.Errorf("expected error field with message, got %v", data["error"])
		}
	})
}

func TestError_String(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		err := New("Simple error", http.StatusBadRequest)
		if err.Error() != "Simple error" {
			t.Errorf("expected 'Simple error', got '%s'", err.Error())
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		baseErr := errors.New("base error")
		err := Wrap(baseErr, "Wrapper message", http.StatusInternalServerError)
		if err.Error() != "Wrapper message: base error" {
			t.Errorf("expected 'Wrapper message: base error', got '%s'", err.Error())
		}
	})

	t.Run("with same message as wrapped error", func(t *testing.T) {
		baseErr := errors.New("same message")
		err := Report(baseErr, http.StatusInternalServerError)
		if err.Error() != "same message" {
			t.Errorf("expected 'same message', got '%s'", err.Error())
		}
	})
}

func TestMergeErrorData(t *testing.T) {
	t.Run("Code and Details merge", func(t *testing.T) {
		code := Code("VALIDATION_ERROR")
		details := Details{
			"email":    "invalid format",
			"password": "too short",
		}
		err := New("Validation failed", http.StatusBadRequest, code, details)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code %s, got %v", code, data["code"])
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if len(detailsMap) != 2 {
			t.Errorf("expected 2 details, got %d", len(detailsMap))
		}
		if detailsMap["email"] != "invalid format" {
			t.Errorf("unexpected email: %s", detailsMap["email"])
		}
	})

	t.Run("later Details replaces earlier", func(t *testing.T) {
		details1 := Details{"field1": "error1"}
		details2 := Details{"field2": "error2"}
		err := New("Multiple details", http.StatusBadRequest, details1, details2)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		detailsMap, ok := data["details"].(Details)
		if !ok {
			t.Fatalf("expected details in data, got %v", data)
		}
		if detailsMap["field1"] == "error1" {
			t.Errorf("expected field1 to be overridden")
		}
		if detailsMap["field2"] != "error2" {
			t.Errorf("expected field2 from second Details, got %v", detailsMap)
		}
	})

	t.Run("later Code overrides earlier", func(t *testing.T) {
		code1 := Code("FIRST")
		code2 := Code("SECOND")
		err := New("Code override", http.StatusBadRequest, code1, code2)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code2 {
			t.Errorf("expected code SECOND (from later Code), got %v", data["code"])
		}
	})

	t.Run("Fields merge with Code and Details", func(t *testing.T) {
		code := Code("ERROR")
		details := Details{"field": "error"}
		fields := Fields{
			"request_id": "123",
			"trace_id":   "456",
		}
		err := New("Test", http.StatusBadRequest, code, details, fields)

		data := err.ErrorData()
		if data == nil {
			t.Fatal("expected error data, got nil")
		}
		if data["code"] != code {
			t.Errorf("expected code in data")
		}
		if _, ok := data["details"]; !ok {
			t.Errorf("expected details in data")
		}
		if data["request_id"] != "123" {
			t.Errorf("expected request_id in data")
		}
		if data["trace_id"] != "456" {
			t.Errorf("expected trace_id in data")
		}
	})
}
