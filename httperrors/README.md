# httperrors

Structured HTTP error handling with status codes and details.

This package provides a way to create errors that carry HTTP status codes and additional context, making it easy to convert application errors into proper HTTP responses.

## Features

- **Status Codes** - Attach HTTP status codes to errors
- **Error Details** - Include field-level validation errors or additional context
- **Error Wrapping** - Wrap existing errors while preserving the error chain
- **Type-Safe Unwrapping** - Extract HTTPError from any error chain

## Installation

```bash
go get github.com/eriicafes/httpx/httperrors
```

## Usage

### Creating Errors

Create new errors with custom messages and status codes. Use this when you're creating domain-specific errors that don't wrap existing errors.

```go
import "github.com/eriicafes/httpx/httperrors"

// Simple HTTP error
// Returns: {"error": "User not found"} with status 404
err := httperrors.New("User not found", http.StatusNotFound)

// With additional details for field-level errors
// Returns: {"error": "Invalid input", "details": {"email": "...", "password": "..."}} with status 400
err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email":    "must be a valid email address",
    "password": "must be at least 8 characters",
})
```

### Reporting Errors

Convert standard Go errors into HTTPErrors while preserving their original error messages. Use `Report` when you want to expose the actual error message to clients (e.g., database errors in development, validation errors).

```go
// Report - converts standard error, using its Error() message
// Use when you want the original error message in the HTTP response
if err := db.QueryRow(...); err != nil {
    // Returns: {"error": "sql: no rows in result set"} with status 500
    return httperrors.Report(err, http.StatusInternalServerError)
}

// With details
// Returns: {"error": "validation failed: ...", "details": {...}} with status 400
err := httperrors.ReportDetails(validationErr, http.StatusBadRequest, httperrors.Details{
    "field": "validation error",
})
```

### Wrapping Errors

Wrap existing errors with custom HTTP-friendly messages while preserving the error chain for logging. Use `Wrap` when you want to hide internal error details from clients but still maintain them for debugging.

```go
// Wrap - add custom message while preserving error chain
// Use when you want a different message than the original error
if err := processPayment(); err != nil {
    // Returns: {"error": "Payment processing failed"} with status 502
    // Original error preserved in error chain for logging
    return httperrors.Wrap(err, "Payment processing failed", http.StatusBadGateway)
}

// Wrap with details
err := httperrors.WrapDetails(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
    "email": "already in use",
})

// When wrapping an HTTPError, empty fields inherit from the wrapped error
baseErr := httperrors.New("Not found", http.StatusNotFound)
err := httperrors.WrapDetails(baseErr, "Resource not found", 0, nil)
// Result: message="Resource not found", statusCode=404 (inherited), details=nil
```

### Unwrapping

Extract HTTPError from the error chain to access status codes and details. This works even if the HTTPError is wrapped in multiple layers of standard Go errors.

```go
// Extract HTTPError from any error in the error chain
if httpErr, ok := httperrors.Unwrap(err); ok {
    log.Printf("HTTP %d: %s", httpErr.StatusCode(), httpErr.Message())
    // Get all details
    message, statusCode, details := httpErr.HTTPError()
}

// Works through multiple layers of standard error wrapping
baseHTTPErr := httperrors.New("Base error", http.StatusNotFound)
wrappedErr := fmt.Errorf("context: %w", baseHTTPErr)
httpErr, ok := httperrors.Unwrap(wrappedErr) // true, finds baseHTTPErr
```

### Accessing Error Information

Access individual fields or get all information at once with automatic defaults for empty values.

```go
err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email": "invalid format",
})

// Individual fields
message := err.Message()       // "Invalid input"
statusCode := err.StatusCode() // 400
details := err.Details()       // map[string]string{"email": "invalid format"}

// All fields at once with automatic defaults
message, statusCode, details := err.HTTPError()
// If message is empty: defaults to "Something went wrong"
// If statusCode is 0: defaults to 500
// If details is nil: returns nil
```

## API

```go
type HTTPError interface {
    error
    Message() string
    StatusCode() int
    Details() Details
    HTTPError() (message string, statusCode int, details Details)
}

type Details = map[string]string

func New(message string, statusCode int) HTTPError
func NewDetails(message string, statusCode int, details Details) HTTPError
func Report(err error, statusCode int) HTTPError
func ReportDetails(err error, statusCode int, details Details) HTTPError
func Wrap(err error, message string, statusCode int) HTTPError
func WrapDetails(err error, message string, statusCode int, details Details) HTTPError
func Unwrap(err error) (HTTPError, bool)
```
