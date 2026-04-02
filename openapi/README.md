# openapi

OpenAPI document generation for `httpx`, with automatic schema reflection and a Scalar reference UI.

It is built on top of `github.com/pb33f/libopenapi` for the OpenAPI document model and `github.com/MarceloPetrucio/go-scalar-api-reference` for the Scalar reference UI.

## Installation

Install the package with `go get`:

```bash
go get github.com/eriicafes/httpx/openapi
```

## Quick Start

This example wires `openapi` into an `httpx` mux, documents a route, and serves openapi JSON and Scalar docs:

```go
package main

import (
	"net/http"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi"
	"github.com/eriicafes/httpx/openapi/doc"
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/resp"
)

type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}

func main() {
	mux := httpx.New()
	mux = openapi.WithRouter(mux, "My API", "1.0.0",
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
			return httpx.Send(w, User{ID: 1})
		},
	)

	mux.Handle("GET /docs", router.OpenAPIJSONHandler())
	mux.Handle("GET /docs/reference", router.ReferenceHandler(nil))

	httpx.ListenAndServe(":8080", mux, nil)
}
```

## Setup

Use `WithRouter` to attach OpenAPI document generation to an `httpx` mux:

```go
mux := httpx.New()
mux = openapi.WithRouter(mux, "User API", "1.0.0",
	doc.Description("API description"),
	doc.Server("https://api.example.com", "Production"),
)

router := openapi.UseRouter(mux)
```

`UseRouter` walks the mux chain, so middleware, prefixes, and other wrapping muxes still apply when you register routes through the returned router.

If you want to build a document before binding to a mux, create a standalone router:

This version records paths first and leaves handler registration for later:

```go
router := openapi.NewRouter("User API", "1.0.0",
	doc.Description("API description"),
)

router.Operation("GET /users",
	op.Summary("List users"),
	op.Response[[]User](200),
)
```

To register handlers later, bind that router to a mux with `WithMux`:

Once bound, the returned router can register handlers just like one created with `WithRouter`:

```go
mux := httpx.New()
r := router.WithMux(mux)
r.Route("POST /users", op.Options(...), handler)
```

## Registering Operations

`Router` supports a few ways to register operations:

Use `Route` when you want to document an operation and register an `httpx` error-returning handler at the same time:

```go
router.Route("GET /health",
	op.Options(
		op.Summary("Health check"),
		op.OperationId("getHealth"),
		op.Response[map[string]string](200, resp.Description("Service is healthy")),
	),
	func(w http.ResponseWriter, r *http.Request) error {
		return httpx.Send(w, map[string]string{"status": "ok"})
	},
)
```

`Router` also supports `Handle` and `HandleFunc` when you already have a standard `http.Handler` or `func(http.ResponseWriter, *http.Request)`.

Use `Operation` when you only want to record the OpenAPI operation and register the runtime handler separately:

```go
router.Operation("GET /health", op.Options(
	op.Summary("Health check"),
	op.Response[map[string]string](200),
))

mux.HandleFunc("GET /health", healthHandler)
```

## Schemas

Go types are reflected automatically:

These are the built-in Go-to-schema mappings used during reflection:

```text
string                -> { "type": "string" }
bool                  -> { "type": "boolean" }
int, int64, uint, ... -> { "type": "integer" }
float32, float64      -> { "type": "number" }
[]T                   -> { "type": "array", "items": <T schema> }
map[K]V               -> { "type": "object", "additionalProperties": <V schema> }
struct                -> { "type": "object", "properties": { ... } }
*T                    -> same as T with "nullable": true
time.Time             -> { "type": "string", "format": "date-time" }
```

Non-pointer struct fields become required by default. Unexported fields and `json:"-"` fields are skipped.

Common schema helpers:

Use these helpers to add titles, examples, enums, field-level overrides, and other general schema metadata:

```go
schema.Options(
	schema.Reference("TypeName"),
	schema.Title("Title"),
	schema.Description("Description"),
	schema.Format("uuid"),
	schema.ReadOnly(),
	schema.WriteOnly(),
	schema.Deprecated(),
	schema.Nullable(),
	schema.Default("value"),
	schema.Example("example"),
	schema.Examples("a", "b"),
	schema.Const("fixed"),
	schema.Enum("a", "b", "c"),
	schema.ExternalDocs("https://docs.example.com", "Description"),
	schema.Field("name", schema.MinLength(1)),
)
```

Array, object, and composition helpers use generics or `schema.Type(...)` getters:

These helpers cover tuple validation, object rules, and schema composition:

```go
schema.Options(
	schema.MinItems(1),
	schema.MaxItems(10),
	schema.UniqueItems(),
	schema.PrefixItems(schema.Type[int](), schema.Type[string]()),
	schema.Contains[MyStruct](),
	schema.UnevaluatedItems[MyStruct](),
	schema.AdditionalProperties[any](),
	schema.DependentSchemas[Extra]("flag"),
	schema.PropertyNames[string](),
	schema.OneOf(schema.Type[Circle](), schema.Type[Square]()),
	schema.AnyOf(schema.Type[A](), schema.Type[B]()),
	schema.AllOf(schema.Type[A](), schema.Type[B]()),
	schema.Not[A](),
	schema.If[A](),
	schema.Then[B](),
	schema.Else[C](),
)
```

A type may set default schema options by implementing the `schema.Schema` interface:

```go
func (User) Schema() schema.Option {
	return schema.Options(
		schema.Reference("User"),
		schema.Field("name", schema.MinLength(3), schema.MaxLength(50)),
		schema.Field("age", schema.Minimum(18), schema.Maximum(120)),
		schema.Field("email", schema.Email(), schema.Example("user@example.com")),
	)
}
```

### Union Types

Types from `github.com/eriicafes/union` expand automatically.

An untagged union expands its cases into `anyOf`:

```go
type Shape union.Union[ShapeSpec]

type ShapeSpec struct {
	Circle *Circle
	Square *Square
}
```

That produces an untagged `anyOf` across `Circle` and `Square`.

Tagged unions expand into `oneOf` and add a discriminator-backed wrapper object for each case:

```go
type Shape union.TaggedUnion[ShapeSpec]
```

That produces a tagged `oneOf` with a discriminator, using `"type"` and `"value"` by default.

## Document Options

Document metadata is configured with `doc.Option` values passed to `WithRouter` or `NewRouter`:

Use `doc` options for top-level metadata, external docs, security requirements, and component registration:

```go
mux = openapi.WithRouter(mux, "User API", "2.0.0",
	doc.Version("3.1.0"),
	doc.Summary("Short summary"),
	doc.Description("Full description of the API"),
	doc.TermsOfService("https://example.com/terms"),
	doc.Contact("API Team", "https://example.com", "api@example.com"),
	doc.License("MIT", "https://opensource.org/licenses/MIT", ""),

	doc.Server("http://localhost:8080", "Local"),
	doc.Server("https://{region}.api.example.com", "Production",
		server.Variable("region", "us-east",
			server.VariableEnum("us-east", "eu-west"),
			server.VariableDescription("Deployment region"),
		),
	),

	doc.Tag("users", "User management"),
	doc.Tag("health", "Service health",
		tag.ExternalDocs("https://docs.example.com/health", "Health check docs"),
	),

	doc.ExternalDocs("https://docs.example.com", "Full documentation"),
	doc.SecurityScheme("bearerAuth",
		securityscheme.HTTP("bearer"),
		securityscheme.BearerFormat("JWT"),
	),
	doc.Security("bearerAuth"),
)
```

### Server Options

Use `server.Option` values to refine `doc.Server(...)` or `pathitem.Server(...)` entries:

```go
doc.Server("https://{region}.api.example.com", "Production",
	server.Name("production"),
	server.Variable("region", "us-east",
		server.VariableEnum("us-east", "eu-west"),
		server.VariableDescription("Deployment region"),
	),
)
```

### Tag Options

Use `tag.Option` values to enrich tags added through `doc.Tag(...)`:

```go
doc.Tag("health", "Service health",
	tag.Summary("Operational endpoints"),
	tag.ExternalDocs("https://docs.example.com/health", "Health check docs"),
)
```

### Security Scheme Options

Use `securityscheme.Option` values when registering reusable schemes with `doc.SecurityScheme(...)`:

```go
doc.SecurityScheme("bearerAuth",
	securityscheme.Description("JWT bearer authentication"),
	securityscheme.HTTP("bearer"),
	securityscheme.BearerFormat("JWT"),
)
```

## Path Item Options

Use `pathitem.Option` values for path-level metadata, shared parameters, and server overrides:

```go
path := pathitem.New(
	pathitem.Summary("Health endpoints"),
	pathitem.Description("Operations for checking service health"),
	pathitem.Parameter[string](param.InHeader, "X-Region"),
	pathitem.Server("https://internal.example.com", "Internal"),
)
```

Use `pathitem.Path` when you want to group multiple methods and shared path-level settings under one route:

```go
path := pathitem.New(
	pathitem.Description("Health endpoints"),
)

path.Route(http.MethodGet,
	op.Options(
		op.Summary("Get health"),
		op.Response[map[string]string](200),
	),
	getHealth,
)

path.Route(http.MethodPost,
	op.Options(
		op.Summary("Run health check"),
		op.Response[map[string]string](200),
	),
	runHealthCheck,
)

router.Path("/health", path)
```

## Operation Options

Combine operation-level settings with `op.Options`:

```go
op.Options(
	op.Summary("Short summary"),
	op.Description("Full description"),
	op.OperationId("uniqueOperationId"),
	op.Tags("users", "admin"),
	op.Deprecated(),
	op.ExternalDocs("https://docs.example.com", "External docs"),
	op.Server("https://api.example.com", "Override server"),
	op.Security("bearerAuth", "scope1"),
	op.PathParam[int]("id"),
	op.QueryParam[string]("search"),
	op.HeaderParam[string]("X-API-Key"),
	op.CookieParam[string]("session"),
	op.RequestBody[CreateUserRequest](),
	op.Response[User](200),
	op.Response[op.NoContent](204),
)
```

## Parameter Options

This example shows how to refine a generated parameter with descriptions, examples, serialization rules, and a component reference:

```go
op.QueryParam[string]("search",
	param.Description("Filter by name or email"),
	param.Required(true),
	param.Deprecated(),
	param.Example("john"),
	param.Style("form"),
	param.Explode(true),
	param.AllowEmptyValue(),
	param.AllowReserved(),
	param.NamedExample("basic", "john"),
	param.Reference("Search"),
)
```

A type may set default parameter options by implementing the `param.Parameter` interface:

```go
func (T) Parameter() param.Option
```

those defaults are applied before inline options.

## Request Bodies

`op.RequestBody[T]` creates a request body for `T` and adds `application/json` by default. Use `op.NoContent` when an operation has no body.

Add request-body metadata, alternate media types, or store the body as a reusable component:

```go
op.RequestBody[CreateUserRequest](
	body.Description("User details"),
	body.Required(true),
	body.ContentType[CreateUserRequest]("application/xml"),
	body.Reference("CreateUser"),
)
```

A type may set default request body options by implementing the `body.RequestBody` interface:

```go
func (T) RequestBody() body.Option
```

those defaults are applied before inline options.

## Media Type Options

Use `mediatype.Option` values to refine generated content entries for request and response bodies:

```go
body.ContentType[CreateUserRequest]("application/json",
	mediatype.Example(CreateUserRequest{Name: "Ada"}),
	mediatype.NamedExample("sample", CreateUserRequest{Name: "Grace"}),
	mediatype.Encoding("avatar", mediatype.EncodingContentType("image/png")),
)
```

A type may set default media type options by implementing the `mediatype.MediaType` interface:

```go
func (T) MediaType() mediatype.Option
```

those defaults are applied before inline options.

## Responses

`op.Response[T](status, ...)` uses `http.StatusText(status)` as the default description and adds `application/json` content unless `T` is `op.NoContent`.

Responses can declare headers, links, alternate media types, and reusable component references:

```go
op.Response[User](200,
	resp.Description("User found"),
	resp.Summary("Successful response"),
	resp.Header[string]("X-Request-Id",
		header.Description("Unique request trace ID"),
		header.Required(true),
	),
	resp.Link("getUser",
		link.OperationId("getUser"),
		link.Parameter("userId", "$response.body#/id"),
		link.Description("Fetch the user by ID"),
	),
	resp.ContentType[User]("application/xml"),
	resp.Reference("UserResponse"),
)
```

A type may set default response options by implementing the `resp.Response` interface:

```go
func (T) Response() resp.Option
```

those defaults are applied before inline options.

## Header Options

Use `header.Option` values for response headers:

```go
resp.Header[string]("X-Request-Id",
	header.Description("Unique request trace ID"),
	header.Style("simple"),
	header.Example("req_123"),
	header.NamedExample("sample", "req_456"),
)
```

A type may set default header options by implementing the `header.Header` interface:

```go
func (T) Header() header.Option
```

those defaults are applied before inline options.

## Link Options

Use `link.Option` values to describe follow-up operations available from a response:

```go
resp.Link("getUser",
	link.OperationId("getUser"),
	link.Parameter("userId", "$response.body#/id"),
	link.Description("Fetch the created user by ID"),
	link.Server("https://api.example.com", "Production"),
)
```

## Example Options

Use `example.Option` values when adding named examples to parameters, headers, or media types:

```go
param.NamedExample("basic", "healthy",
	example.Summary("Basic health state"),
	example.Description("Returned when the service is healthy"),
)
```

## Callback Options

Use `callback.Option` values to attach callbacks to operations:

```go
op.Options(
	op.Callback("onHealthChange",
		callback.Reference("HealthStatusChanged"),
	),
)
```

## Advanced Patterns

A single type may implement more than one interface at a time. This lets one type provide defaults at multiple OpenAPI layers.

This is especially useful with `Reference(...)` options. A type can set its own component reference through an interface-provided default, which keeps call sites short and reuses the same component everywhere.

For example, a response type can set default schema, response and media type options at the same time:

```go
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (User) Schema() schema.Option {
	return schema.Options(
		// Register the reflected schema in components/schemas.
		schema.Reference("User"),
		schema.Field("name", schema.MinLength(1)),
	)
}

func (User) Response() resp.Option {
	return resp.Options(
		// Register the response in components/responses.
		resp.Reference("UserResponse"),
		resp.Header[string]("X-Resource-Type",
			header.Example("user"),
		),
	)
}

func (User) MediaType() mediatype.Option {
	return mediatype.Options(
		mediatype.NamedExample("default", User{
			ID:   1,
			Name: "Ada",
		}),
	)
}

op.Response[User](200)
```

Likewise, a request type can combine schema, request body, and media type defaults:

```go
type CreateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

func (CreateUserRequest) Schema() schema.Option {
	return schema.Options(
		schema.Reference("CreateUserRequest"),
		schema.Field("email", schema.Email()),
	)
}

func (CreateUserRequest) RequestBody() body.Option {
	return body.Options(
		body.Description("Payload for creating a user"),
		body.Reference("CreateUserBody"),
	)
}

func (CreateUserRequest) MediaType() mediatype.Option {
	return mediatype.Options(
		mediatype.NamedExample("default", CreateUserRequest{
			Name:  "Ada",
			Email: "ada@example.com",
		}),
	)
}

op.RequestBody[CreateUserRequest]()
```

References also work inline. One call site can define the full object, and another can reuse it by reference:

```go
// First call site: define the reusable response and store it in components/responses.
op.Response[User](200,
	resp.Description("User response"),
	resp.Header[string]("X-Request-Id",
		header.Description("Unique request trace ID"),
	),
	resp.Reference("UserResponse"),
)

// Later call site: reuse the stored component by reference.
op.Response[User](200,
	resp.Reference("UserResponse"),
)
```

Inline options are skipped only when the type itself sets `Reference(...)` through an implemented interface. That keeps the stored component unambiguous even when different call sites pass different inline options.

## Serving the Document

Serve the generated document directly as JSON or YAML:

```go
mux.Handle("GET /docs/openapi.json", router.OpenAPIJSONHandler())
mux.Handle("GET /docs/openapi.yaml", router.OpenAPIYAMLHandler())
```

To serve Scalar:

Use the built-in Scalar handler when the default embedded configuration is enough:

```go
mux.Handle("GET /docs/reference", router.ReferenceHandler(nil))
```

Or with custom options:

Pass `scalar.Options` to point Scalar at a hosted spec URL or customize the page:

```go
mux.Handle("GET /docs", router.ReferenceHandler(&scalar.Options{
	SpecURL: "http://localhost:8080/docs/openapi.json",
	Theme:   scalar.ThemeDefault,
	CustomOptions: scalar.CustomOptions{
		PageTitle: "My API Reference",
	},
}))
```

If you need the raw document:

You can always access the underlying `*v3.Document` from `github.com/pb33f/libopenapi/datamodel/high/v3` for custom rendering or validation:

```go
doc := router.GetDocument()
yaml, err := doc.Render()
json, err := doc.RenderJSON("")
```
