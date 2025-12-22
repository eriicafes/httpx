# ContextKey

### Type-safe context key management for Go using generics.

ContextKey provides type-safe context key implementations with compile-time type safety, eliminating the need for type assertions and reducing runtime errors when working with Go's context values.

## Installation

```sh
go get github.com/eriicafes/httpx
```

## Usage

### Define a context key

Create a context key with a specific type using `New`. The type parameter determines what values can be stored.

```go
package main

import (
    "context"
    "github.com/eriicafes/httpx/contextkey"
)

var (
    userIDKey    = contextkey.New[int]("userID")
    requestIDKey = contextkey.New[string]("requestID")
    userKey      = contextkey.New[*User]("user")
)

type User struct {
    ID   int
    Name string
}
```

### Set and get values

```go
func main() {
    ctx := context.Background()

    // Set a value
    ctx = userIDKey.Set(ctx, 12345)

    // Get a value
    if id, ok := userIDKey.Get(ctx); ok {
        fmt.Println("User ID:", id) // User ID: 12345
    }
}
```

### Delete values

```go
ctx = userIDKey.Set(ctx, 12345)

// Remove the value
ctx = userIDKey.Delete(ctx)

if _, ok := userIDKey.Get(ctx); !ok {
    fmt.Println("User ID not found")
}
```

### Take values (get and delete)

Take retrieves a value and removes it from the context in a single operation.

```go
ctx = requestIDKey.Set(ctx, "abc-123")

// Get and remove in one operation
newCtx, requestID, ok := requestIDKey.Take(ctx)
if ok {
    fmt.Println("Request ID:", requestID) // Request ID: abc-123
}

// Value is now removed from newCtx
if _, ok := requestIDKey.Get(newCtx); !ok {
    fmt.Println("Request ID no longer in context")
}
```

### Update values

Update modifies an existing value or sets a new one using an update function.

```go
ctx = userIDKey.Set(ctx, 100)

// Increment the user ID
ctx = userIDKey.Update(ctx, func(prev int, hasPrev bool) int {
    if hasPrev {
        return prev + 1
    }
    return 1 // Default value if not set
})

if id, ok := userIDKey.Get(ctx); ok {
    fmt.Println("Updated ID:", id) // Updated ID: 101
}
```

## Benefits

- **Type safety**: Compile-time type checking prevents type mismatches
- **No type assertions**: Values are returned with the correct type automatically
- **Clear API**: Explicit methods for common operations (Get, Set, Delete, Take, Update)
- **Zero dependencies**: Uses only the standard library
- **Lightweight**: Minimal overhead over native context operations
