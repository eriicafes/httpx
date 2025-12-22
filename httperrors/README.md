# httperrors

Structured HTTP error handling with status codes and details.

## Usage

### Creating Errors

```go
// Simple HTTP error
err := httperrors.New("User not found", http.StatusNotFound)

// With additional details
err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email":    "must be a valid email address",
    "password": "must be at least 8 characters",
})
```

### Converting Standard Errors

```go
// Report a standard error as HTTP error
if err := db.QueryRow(...); err != nil {
    return httperrors.Report(err, http.StatusInternalServerError)
}

// With details
err := httperrors.ReportDetails(err, http.StatusBadRequest, httperrors.Details{
    "field": "validation error",
})
```

### Wrapping Errors

```go
// Wrap with custom message
err := httperrors.Wrap(err, "Payment processing failed", http.StatusBadGateway)

// Wrap with details
err := httperrors.WrapDetails(err, "Validation failed", http.StatusBadRequest, httperrors.Details{
    "email": "already in use",
})

// When wrapping an HTTPError, empty fields inherit from the wrapped error
baseErr := httperrors.New("Not found", http.StatusNotFound)
err := httperrors.WrapDetails(baseErr, "Resource not found", 0, nil) // inherits 404 status
```

### Unwrapping

```go
// Extract HTTPError from error chain
if httpErr, ok := httperrors.Unwrap(err); ok {
    log.Printf("HTTP %d: %s", httpErr.StatusCode(), httpErr.Message())
}

// Works through multiple layers of standard error wrapping
baseHTTPErr := httperrors.New("Base error", http.StatusNotFound)
err := fmt.Errorf("wrapper: %w", baseHTTPErr)
httpErr, ok := httperrors.Unwrap(err) // finds baseHTTPErr
```

### Accessing Error Information

```go
err := httperrors.NewDetails("Invalid input", http.StatusBadRequest, httperrors.Details{
    "email": "invalid format",
})

// Individual fields
message := err.Message()
statusCode := err.StatusCode()
details := err.Details()

// All fields with defaults
message, statusCode, details := err.HTTPError()
// Returns "Something went wrong" and 500 for empty values
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
