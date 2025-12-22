package httperrors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestNew(t *testing.T) {
	err := New("User not found", http.StatusNotFound)

	if err.Message() != "User not found" {
		t.Errorf("expected message 'User not found', got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, err.StatusCode())
	}
	if err.Details() != nil {
		t.Errorf("expected nil details, got %v", err.Details())
	}
	if err.Error() != "User not found" {
		t.Errorf("expected error string 'User not found', got '%s'", err.Error())
	}
}

func TestNewDetails(t *testing.T) {
	details := Details{
		"email":    "must be a valid email address",
		"password": "must be at least 8 characters",
	}
	err := NewDetails("Invalid input", http.StatusBadRequest, details)

	if err.Message() != "Invalid input" {
		t.Errorf("expected message 'Invalid input', got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
	}
	if len(err.Details()) != 2 {
		t.Errorf("expected 2 details, got %d", len(err.Details()))
	}
	if err.Details()["email"] != "must be a valid email address" {
		t.Errorf("unexpected email detail: %s", err.Details()["email"])
	}
	if err.Details()["password"] != "must be at least 8 characters" {
		t.Errorf("unexpected password detail: %s", err.Details()["password"])
	}
}

func TestReport(t *testing.T) {
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
}

func TestReportDetails(t *testing.T) {
	baseErr := errors.New("validation failed")
	details := Details{"username": "required field"}
	err := ReportDetails(baseErr, http.StatusBadRequest, details)

	if err.Message() != "validation failed" {
		t.Errorf("expected message 'validation failed', got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
	}
	if err.Details()["username"] != "required field" {
		t.Errorf("unexpected username detail: %s", err.Details()["username"])
	}
	if !errors.Is(err, baseErr) {
		t.Error("expected error to unwrap to base error")
	}
}

func TestWrap(t *testing.T) {
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
}

func TestWrapDetails(t *testing.T) {
	baseErr := errors.New("validation error")
	details := Details{"email": "already in use"}
	err := WrapDetails(baseErr, "Validation failed", http.StatusBadRequest, details)

	if err.Message() != "Validation failed" {
		t.Errorf("expected message 'Validation failed', got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
	}
	if err.Details()["email"] != "already in use" {
		t.Errorf("unexpected email detail: %s", err.Details()["email"])
	}
	if err.Error() != "Validation failed: validation error" {
		t.Errorf("unexpected error string: %s", err.Error())
	}
}

func TestWrapHTTPError_InheritStatusCode(t *testing.T) {
	baseErr := New("Original error", http.StatusNotFound)
	err := WrapDetails(baseErr, "Wrapped error", 0, nil)

	if err.StatusCode() != http.StatusNotFound {
		t.Errorf("expected status code %d (inherited), got %d", http.StatusNotFound, err.StatusCode())
	}
	if err.Message() != "Wrapped error" {
		t.Errorf("expected message 'Wrapped error', got '%s'", err.Message())
	}
}

func TestWrapHTTPError_InheritMessage(t *testing.T) {
	baseErr := New("Original message", http.StatusNotFound)
	err := WrapDetails(baseErr, "", http.StatusBadRequest, nil)

	if err.Message() != "Original message" {
		t.Errorf("expected message 'Original message' (inherited), got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
	}
}

func TestWrapHTTPError_InheritDetails(t *testing.T) {
	baseDetails := Details{"field": "error"}
	baseErr := NewDetails("Original", http.StatusBadRequest, baseDetails)
	err := WrapDetails(baseErr, "Wrapped", http.StatusBadRequest, nil)

	if err.Details()["field"] != "error" {
		t.Errorf("expected inherited details, got %v", err.Details())
	}
}

func TestWrapHTTPError_OverrideAll(t *testing.T) {
	baseErr := NewDetails("Original", http.StatusNotFound, Details{"old": "value"})
	newDetails := Details{"new": "value"}
	err := WrapDetails(baseErr, "New message", http.StatusBadRequest, newDetails)

	if err.Message() != "New message" {
		t.Errorf("expected message 'New message', got '%s'", err.Message())
	}
	if err.StatusCode() != http.StatusBadRequest {
		t.Errorf("expected status code %d, got %d", http.StatusBadRequest, err.StatusCode())
	}
	if err.Details()["new"] != "value" {
		t.Errorf("expected new details, got %v", err.Details())
	}
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
		baseHTTPErr := NewDetails("HTTP error", http.StatusBadRequest, Details{"field": "error"})
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
		if httpErr.Details()["field"] != "error" {
			t.Errorf("expected details to be preserved, got %v", httpErr.Details())
		}
	})
}

func TestHTTPError_Method(t *testing.T) {
	t.Run("with all values", func(t *testing.T) {
		details := Details{"field": "error"}
		err := NewDetails("Test message", http.StatusBadRequest, details)
		message, statusCode, returnedDetails := err.HTTPError()

		if message != "Test message" {
			t.Errorf("expected message 'Test message', got '%s'", message)
		}
		if statusCode != http.StatusBadRequest {
			t.Errorf("expected status code %d, got %d", http.StatusBadRequest, statusCode)
		}
		if returnedDetails["field"] != "error" {
			t.Errorf("unexpected details: %v", returnedDetails)
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

	t.Run("with empty message and zero status code", func(t *testing.T) {
		err := New("", 0)
		message, statusCode, _ := err.HTTPError()

		if message != "Something went wrong" {
			t.Errorf("expected default message 'Something went wrong', got '%s'", message)
		}
		if statusCode != http.StatusInternalServerError {
			t.Errorf("expected default status code %d, got %d", http.StatusInternalServerError, statusCode)
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
