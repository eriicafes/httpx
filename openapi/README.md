# openapi

OpenAPI document generation with automatic schema reflection and a Scalar reference UI.

## Installation

```bash
go get github.com/eriicafes/httpx/openapi
```

## Dependencies

- [**pb33f/libopenapi**](https://github.com/pb33f/libopenapi) - OpenAPI document model and JSON/YAML rendering
- [**MarceloPetrucio/go-scalar-api-reference**](https://github.com/MarceloPetrucio/go-scalar-api-reference) - Scalar API reference UI
- [**eriicafes/union**](https://github.com/eriicafes/union) - Union types for `allOf`, `oneOf`, and `anyOf` schema types (optional)

## Subpackages

- **doc** - Document options
- **op** - Operation options
- **param** - Parameter options
- **body** - Request body options
- **resp** - Response options
- **header** - Response header options
- **schema** - JSON Schema options

## Quick Start

```go
import (
    "net/http"

    "github.com/eriicafes/httpx"
    "github.com/eriicafes/httpx/openapi"
    "github.com/eriicafes/httpx/openapi/doc"
    "github.com/eriicafes/httpx/openapi/op"
    "github.com/eriicafes/httpx/openapi/resp"
)

mux := httpx.New()
mux = openapi.WithRouter(mux,
    doc.Info("My API", "1.0.0"),
    doc.Server("http://localhost:8080", "Local"),
)

router := openapi.UseRouter(mux)

router.Route("GET /users/{id}",
    op.Options(
        op.Summary("Get user"),
        op.PathParam[int]("id"),
        op.Response[User](200, resp.Description("User found")),
        op.Response[ErrorResponse](404, resp.Description("User not found")),
    ),
    func(w http.ResponseWriter, r *http.Request) error {
        return httpx.Send(w, User{})
    },
)

// Serve the OpenAPI spec and Scalar API reference UI
mux.Handle("GET /docs", router.OpenAPIHandler())
mux.Handle("GET /docs/reference", router.ReferenceHandler(nil))

httpx.ListenAndServe(":8080", mux, nil)
```

## Setup

### WithRouter

Attach an OpenAPI router to an existing mux:

```go
mux := httpx.New()
mux = openapi.WithRouter(mux,
    doc.Info("My API", "1.0.0"),
    doc.Description("API description"),
    doc.Server("https://api.example.com", "Production"),
)

// Extract the router to register routes
router := openapi.UseRouter(mux)
```

`WithRouter` wraps the mux and embeds a `Router`. `UseRouter` traverses the mux chain to find and return it.

### Standalone Router

`NewRouter` creates a router with no mux bound to it. Use `router.Path` to record OpenAPI operations and serve the document without registering any handlers:

```go
router := openapi.NewRouter(
    doc.Info("My API", "1.0.0"),
)

// Record operations only — no handlers registered
router.Path("GET /users", op.Options(...))
router.Path("POST /users", op.Options(...))

http.Handle("GET /docs", router.OpenAPIHandler())
```

To register handlers, bind the router to a mux first with `WithMux`:

```go
mux := http.NewServeMux()
r := router.WithMux(mux)
r.Route("GET /users", op.Options(...), handler)
```

## Document Configuration

Configure the OpenAPI document with `doc` options passed to `WithRouter` or `NewRouter`:

```go
mux = openapi.WithRouter(mux,
    doc.Info("User API", "2.0.0"),
    doc.Summary("Short summary"),
    doc.Description("Full description of the API"),
    doc.TermsOfService("https://example.com/terms"),
    doc.Contact("API Team", "https://example.com", "api@example.com"),
    doc.License("MIT", "https://opensource.org/licenses/MIT"),
    // or with SPDX identifier:
    doc.LicenseIdentifier("MIT", "MIT"),

    doc.Server("http://localhost:8080", "Local"),
    doc.Server("https://api.example.com", "Production",
        doc.ServerVariable("region", "us-east", "us-east", "eu-west"),
    ),

    doc.Tag("users", "User management"),
    doc.Tag("health", "Service health",
        doc.TagExternalDocs("https://docs.example.com/health", "Health check docs"),
    ),

    doc.ExternalDocs("https://docs.example.com", "Full documentation"),

    // Global security requirement
    doc.Security("bearerAuth"),
    // Make security optional globally (empty requirement)
    doc.Security(""),
)
```

## Operation Configuration

Use `op.Options` to combine multiple operation options, then pass it to `router.Route` or `router.Path`:

### router.Route

Registers a route handler and records the OpenAPI operation:

```go
router.Route("POST /users",
    op.Options(
        op.Summary("Create user"),
        op.Tags("users"),
        op.OperationId("createUser"),
        op.RequestBody[CreateUserRequest](body.Required()),
        op.Response[User](201, resp.Description("User created")),
    ),
    func(w http.ResponseWriter, r *http.Request) error {
        // handle request
        w.WriteHeader(http.StatusCreated)
        return httpx.Send(w, User{})
    },
)
```

### router.Path

Records only the OpenAPI operation without registering a handler. Useful when the handler is registered elsewhere:

```go
router.Path("GET /health", op.Options(
    op.Summary("Health check"),
    op.Tags("health"),
    op.Response[map[string]string](200),
))

// Handler registered separately
mux.HandleFunc("GET /health", healthHandler)
```

### Operation Options

```go
op.Options(
    op.Summary("Short summary"),
    op.Description("Full description"),
    op.OperationId("uniqueOperationId"),
    op.Tags("users", "admin"),
    op.Deprecated(),
    op.ExternalDocs("https://docs.example.com", "External docs"),
    op.Server("https://api.example.com", "Override server for this operation"),
    op.Security("bearerAuth", "scope1"),  // pass "" to make security optional
    op.PathParam[int]("id"),              // required; type inferred from T
    op.QueryParam[string]("search"),      // optional
    op.HeaderParam[string]("X-API-Key"), // optional
    op.CookieParam[string]("session"),   // optional
    op.RequestBody[CreateUserRequest](...),
    op.Response[User](200, ...),
    op.Response[op.NoContent](204, ...),  // no response body
)
```

## Parameters

Refine parameters with `param` options:

```go
op.QueryParam[string]("search",
    param.Description("Filter by name or email"),
    param.Required(),
    param.Deprecated(),
    param.Example("john"),
    param.Style("form"),
    param.Explode(true),
    param.AllowEmptyValue(),
    param.AllowReserved(),
    param.NamedExample("basic", "john"),
    param.Reference("#/components/parameters/Search"),
)
```

## Request Body

Refine the request body with `body` options:

```go
op.RequestBody[CreateUserRequest](
    body.Description("User details"),
    body.Required(),
    // Add additional content types alongside application/json
    body.ContentType[CreateUserRequest]("application/xml"),
    body.Reference("#/components/requestBodies/CreateUser"),
)
```

## Responses

Refine responses with `resp` options:

```go
op.Response[User](200,
    resp.Description("User found"),
    resp.Summary("Successful response"),
    resp.Header[string]("X-Request-Id",
        header.Description("Unique request trace ID"),
        header.Required(),
    ),
    resp.Header[int]("X-Rate-Limit-Remaining",
        header.Description("Remaining requests in the current window"),
    ),
    resp.Link("getUser",
        resp.LinkOperationId("getUser"),
        resp.LinkParameter("userId", "$response.body#/id"),
        resp.LinkDescription("Fetch the user by ID"),
    ),
    // Add additional content types alongside application/json
    resp.ContentType[User]("application/xml"),
    resp.Reference("#/components/responses/UserResponse"),
)
```

## Schema Customization

### SchemaOption Interface

Implement `Schema() schema.Option` on your types to customize how they are reflected:

```go
type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

func (u User) Schema() schema.Option {
    return schema.Options(
        // Register in components/schemas as "User" and return a $ref
        schema.Ref("User"),
        // Apply constraints to individual fields
        schema.Field("name", schema.MinLength(3), schema.MaxLength(50)),
        schema.Field("age", schema.Minimum(18), schema.Maximum(120)),
        schema.Field("email", schema.Email(), schema.Example("user@example.com")),
    )
}
```

Types that declare `schema.Ref` are automatically registered in `components/schemas` and referenced inline via `$ref`.

### Schema Options Reference

**General**

```go
schema.Options(
    schema.Ref("TypeName"),         // register in components/schemas, return $ref
    schema.Title("Title"),
    schema.Description("Description"),
    schema.Format("uuid"),          // prefer named helpers: Email(), UUID(), DateTime()
    schema.ReadOnly(),              // value must not be sent in requests
    schema.WriteOnly(),             // value must not be included in responses
    schema.Deprecated(),
    schema.Nullable(),              // prefer pointer types in OpenAPI 3.1
    schema.Default("value"),
    schema.Example("example"),
    schema.Examples("a", "b"),      // OpenAPI 3.1
    schema.Const("fixed"),
    schema.Enum("a", "b", "c"),
    schema.ExternalDocs("https://docs.example.com", "Description"),
    schema.Field("name", ...),      // apply options to a named struct property
)
```

**String**

```go
schema.Options(
    schema.MinLength(3),
    schema.MaxLength(50),
    schema.Pattern(`^\d+$`),
    schema.Email(),
    schema.UUID(),
    schema.DateTime(),
    schema.ContentEncoding("base64"),
    schema.ContentMediaType("image/png"),
)
```

**Numeric**

```go
schema.Options(
    schema.Minimum(0),
    schema.Maximum(100),
    schema.ExclusiveMinimum(0),   // OpenAPI 3.1
    schema.ExclusiveMaximum(100), // OpenAPI 3.1
    schema.MultipleOf(5),
)
```

**Array**

```go
schema.Options(
    schema.MinItems(1),
    schema.MaxItems(10),
    schema.UniqueItems(),
    schema.PrefixItems(int(0), ""),      // tuple validation per position, OpenAPI 3.1
    schema.Contains(MyStruct{}),         // at least one item must match, OpenAPI 3.1
    schema.MinContains(1),
    schema.MaxContains(5),
    schema.UnevaluatedItems(MyStruct{}), // schema for items beyond prefixItems, OpenAPI 3.1
)
```

**Object**

```go
schema.Options(
    schema.MinProperties(1),
    schema.MaxProperties(10),
    schema.AdditionalProperties[any](),                   // use a concrete type to constrain
    schema.PatternProperties(`^\w+$`, schema.Title("...")),
    schema.DependentRequired("flag", "field1", "field2"), // OpenAPI 3.1
    schema.DependentSchemas("flag", Extra{}),             // OpenAPI 3.1
    schema.PropertyNames(""),                             // OpenAPI 3.1
    schema.UnevaluatedProperties[any](),                  // OpenAPI 3.1
)
```

**Composition**

```go
schema.Options(
    schema.AllOf(TypeA{}, TypeB{}),
    schema.OneOf(TypeA{}, TypeB{}),
    schema.AnyOf(TypeA{}, TypeB{}),
    schema.Not(TypeA{}),
    schema.If(TypeA{}),   // OpenAPI 3.1
    schema.Then(TypeB{}), // OpenAPI 3.1
    schema.Else(TypeC{}), // OpenAPI 3.1
)
```

**XML**

```go
schema.Options(
    schema.XMLName("element"),
    schema.XMLNamespace("https://example.com/ns"),
    schema.XMLPrefix("ns"),
    schema.XMLAttribute(), // render as XML attribute instead of element
    schema.XMLWrapped(),   // wrap array in an enclosing XML element
)
```

### Type Reflection

Go types are automatically reflected to JSON Schema:

```
string                → { "type": "string" }
bool                  → { "type": "boolean" }
int, int64, uint, ... → { "type": "integer" }
float32, float64      → { "type": "number" }
[]T                   → { "type": "array", "items": <T schema> }
map[K]V               → { "type": "object", "additionalProperties": <V schema> }
struct                → { "type": "object", "properties": { ... } }
*T                    → same as T with "nullable": true
time.Time             → { "type": "string", "format": "date-time" }
```

Non-pointer struct fields are added to `required`. Unexported fields and `json:"-"` tagged fields are skipped.

### Union Types

Types from `github.com/eriicafes/union` are expanded to `oneOf` automatically.

`union.Union[Spec]` expands to an untagged `oneOf` across all case types in the spec:

```go
type Shape union.Union[ShapeSpec]
type ShapeSpec struct {
    Circle  Circle
    Square  Square
}
// Generates: oneOf: [Circle schema, Square schema]
```

`union.TaggedUnion[Spec]` expands to a tagged `oneOf` with a discriminator. Each case is an object with a `type` key and a `value` key:

```go
type Shape union.TaggedUnion[ShapeSpec]
// Generates: oneOf with discriminator on "type":
//   { "type": "Circle", "value": { ...Circle schema } }
//   { "type": "Square", "value": { ...Square schema } }
```

## Serving the Spec

### OpenAPI JSON

```go
mux.Handle("GET /docs", router.OpenAPIHandler())
// GET /docs returns the OpenAPI JSON spec
```

### Scalar API Reference UI

```go
// Minimal setup — uses the document title and embedded spec
mux.Handle("GET /docs/reference", router.ReferenceHandler(nil))

// Custom Scalar options
mux.Handle("GET /docs/reference", router.ReferenceHandler(&scalar.Options{
    SpecURL: "http://localhost:8080/docs",
    Theme:   scalar.ThemeDefault,
    CustomOptions: scalar.CustomOptions{
        PageTitle: "My API Reference",
    },
}))
```

### Raw Document

Access the underlying `*v3.Document` for custom rendering or validation:

```go
doc := router.GetDocument()
yaml, err := doc.Render()
```

## Composing with httpx Mux

`openapi.WithRouter` returns a standard `httpx.Mux`, so it composes naturally with other httpx mux types:

```go
mux := httpx.New()
mux = httpx.NormalizeTrailingSlash(mux)
mux = openapi.WithRouter(mux, doc.Info("My API", "1.0.0"))
mux = httpx.Use(mux, loggingMiddleware, authMiddleware)
mux = httpx.Fallback(mux, errorHandler)

router := openapi.UseRouter(mux)

router.Route("GET /users", op.Options(
    op.Summary("List users"),
    op.Response[[]User](200),
), func(w http.ResponseWriter, r *http.Request) error {
    // trailing slash normalized, middleware applied, custom error handling
    return httpx.Send(w, []User{})
})
```
