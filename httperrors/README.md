# httperrors

Structured HTTP error handling with status codes and metadata.

## Installation

```bash
go get github.com/eriicafes/httpx
```

## Features

- **Status Codes** - Attach HTTP status codes to errors
- **Error Metadata** - Include structured metadata using `Code`, `Details`, or custom `Fields`
- **Error Wrapping** - Wrap existing errors while preserving the error chain
- **Error Unwrapping** - Extract HTTPError from any error chain

## Usage

### Creating Errors

Create new errors with custom messages and status codes.

```go
import "github.com/eriicafes/httpx/httperrors"

// Simple error
err := httperrors.New("User not found", http.StatusNotFound)

// With validation details
err := httperrors.New("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email":    "must be a valid email address",
    "password": "must be at least 8 characters",
})

// With error code
const CodeNotFound = httperrors.Code("NOT_FOUND")
err := httperrors.New("User not found", http.StatusNotFound, CodeNotFound)

// With custom fields
err := httperrors.New("Internal error", http.StatusInternalServerError, httperrors.Fields{
    "request_id": "abc123",
    "trace_id":   "xyz789",
})

// Mix multiple metadata types
err := httperrors.New("Validation failed", http.StatusBadRequest,
    httperrors.Code("VALIDATION_ERROR"),
    httperrors.Details{"email": "invalid format"},
    httperrors.Fields{"request_id": "req-123"},
)
```

### Reporting Errors

Convert standard Go errors into HTTPErrors while preserving their original error messages. Use `Report` when you want to expose the error message to clients.

```go
// Simple report
if err := db.QueryRow(...); err != nil {
    return httperrors.Report(err, http.StatusInternalServerError)
}

// With additional metadata
if err := validator.Validate(input); err != nil {
    return httperrors.Report(err, http.StatusBadRequest, httperrors.Details{
        "username": "required field",
    })
}
```

### Wrapping Errors

Wrap existing errors with custom user-friendly messages while preserving the error chain for logging. Use `Wrap` when you want to hide internal error details from clients but still maintain them for debugging.

```go
// Simple wrap
if err := processPayment(); err != nil {
    return httperrors.Wrap(err, "Payment processing failed", http.StatusBadGateway)
}

// With metadata
if err := validateUser(input); err != nil {
    return httperrors.Wrap(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
        "email": "already in use",
    })
}

// Inherit from wrapped HTTPError
baseErr := httperrors.New("Not found", http.StatusNotFound)
err := httperrors.Wrap(baseErr, "Resource not found", 0) // Inherits status code 404
```

### Unwrapping

Extract HTTPError from the error chain to access status codes and metadata. This works even if the HTTPError is wrapped in multiple layers of standard Go errors.

```go
if httpErr, ok := httperrors.Unwrap(err); ok {
    log.Printf("HTTP %d: %s", httpErr.StatusCode(), httpErr.Message())
    data := httpErr.ErrorData()
}

// Works through multiple wrappers
baseHTTPErr := httperrors.New("Base error", http.StatusNotFound)
wrappedErr := fmt.Errorf("context: %w", baseHTTPErr)
httpErr, ok := httperrors.Unwrap(wrappedErr) // Returns base error
```

### Accessing Error Information

Access individual fields or get all information at once with automatic defaults for empty values.

```go
err := httperrors.New("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email": "invalid format",
})

// Access individual fields
message := err.Message()       // "Invalid input"
statusCode := err.StatusCode() // 400
data := err.ErrorData()        // map[string]any{"details": Details{"email": "invalid format"}}

// Get everything with defaults
message, statusCode, data := err.HTTPError()
// Returns defaults for empty values:
// - message: "Something went wrong" if empty
// - statusCode: 500 if zero
// - data: includes "error" field with the message
```

## Error Data

Errors can include additional structured metadata beyond the message and status code. This metadata is accessible via the `ErrorData()` method and can be used to provide extra context like error codes, validation details.

### Code

Error codes for standardized error classification. Usually defined as constants.

```go
const (
    CodeNotFound     = httperrors.Code("NOT_FOUND")
    CodeValidation   = httperrors.Code("VALIDATION_ERROR")
    CodeUnauthorized = httperrors.Code("UNAUTHORIZED")
)

err := httperrors.New("User not found", http.StatusNotFound, CodeNotFound)
// err.ErrorData() returns: {"code": "NOT_FOUND"}
```

### Details

Field-level validation errors where specific fields have specific error messages.

```go
details := httperrors.Details{
    "email":    "must be a valid email address",
    "password": "must be at least 8 characters",
}
err := httperrors.New("Invalid input", http.StatusBadRequest, details)
// err.ErrorData() returns: {"details": {"email": "...", "password": "..."}}
```

### Fields

Arbitrary custom fields for any top-level metadata.

```go
fields := httperrors.Fields{
    "request_id": "abc123",
    "trace_id":   "xyz789",
}
err := httperrors.New("Error", http.StatusInternalServerError, fields)
// err.ErrorData() returns: {"request_id": "abc123", "trace_id": "xyz789"}
```

### Combining Multiple Error Data

When multiple `ErrorData` types are provided, they are merged together. Later values override earlier ones for the same keys.

```go
code := httperrors.Code("VALIDATION_ERROR")
details := httperrors.Details{"email": "invalid format"}
fields := httperrors.Fields{"request_id": "req-123"}

err := httperrors.New("Validation failed", http.StatusBadRequest, code, details, fields)
// err.ErrorData() returns merged data:
// {
//   "code": "VALIDATION_ERROR",
//   "details": {"email": "invalid format"},
//   "request_id": "req-123"
// }
```

## API

```go
type HTTPError interface {
    error
    ErrorData
    Message() string
    StatusCode() int
    HTTPError() (message string, statusCode int, data map[string]any)
}

type ErrorData interface {
    ErrorData() map[string]any
}

type Details = map[string]string
type Code = string
type Fields = map[string]any

func New(message string, statusCode int, data ...ErrorData) HTTPError
func Report(err error, statusCode int, data ...ErrorData) HTTPError
func Wrap(err error, message string, statusCode int, data ...ErrorData) HTTPError
func Unwrap(err error) (HTTPError, bool)
```
