# httperrors

Structured HTTP error handling with status codes and details.

## Installation

```bash
go get github.com/eriicafes/httpx
```

## Features

- **Status Codes** - Attach HTTP status codes to errors
- **Error Details** - Include field-level errors for additional context
- **Error Wrapping** - Wrap existing errors while preserving the error chain
- **Error Unwrapping** - Extract HTTPError from any error chain

## Usage

### Creating Errors

Create new errors with custom messages and status codes.

```go
import "github.com/eriicafes/httpx/httperrors"

err := httperrors.New("User not found", http.StatusNotFound)

err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email":    "must be a valid email address",
    "password": "must be at least 8 characters",
})
```

### Reporting Errors

Convert standard Go errors into HTTPErrors while preserving their original error messages. Use `Report` when you want to expose the error message to clients.

```go
if err := db.QueryRow(...); err != nil {
    return httperrors.Report(err, http.StatusInternalServerError)
}

err := httperrors.ReportDetails(validationErr, http.StatusBadRequest, httperrors.Details{
    "field": "validation error",
})
```

### Wrapping Errors

Wrap existing errors with custom user-friendly messages while preserving the error chain for logging. Use `Wrap` when you want to hide internal error details from clients but still maintain them for debugging.

```go
if err := processPayment(); err != nil {
    return httperrors.Wrap(err, "Payment processing failed", http.StatusBadGateway)
}

err := httperrors.WrapDetails(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
    "email": "already in use",
})

baseErr := httperrors.New("Not found", http.StatusNotFound)
err := httperrors.WrapDetails(baseErr, "Resource not found", 0, nil)
```

### Unwrapping

Extract HTTPError from the error chain to access status codes and details. This works even if the HTTPError is wrapped in multiple layers of standard Go errors.

```go
if httpErr, ok := httperrors.Unwrap(err); ok {
    log.Printf("HTTP %d: %s", httpErr.StatusCode(), httpErr.Message())
    message, statusCode, details := httpErr.HTTPError()
}

baseHTTPErr := httperrors.New("Base error", http.StatusNotFound)
wrappedErr := fmt.Errorf("context: %w", baseHTTPErr)
httpErr, ok := httperrors.Unwrap(wrappedErr)
```

### Accessing Error Information

Access individual fields or get all information at once with automatic defaults for empty values.

```go
err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email": "invalid format",
})

message := err.Message()
statusCode := err.StatusCode()
details := err.Details()

message, statusCode, details := err.HTTPError()
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
