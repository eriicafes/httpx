# OpenAPI Reference

Use `github.com/eriicafes/httpx/openapi` when routes should generate an OpenAPI document alongside runtime handler registration.

## Setup

Attach a router to an `httpx` mux with `openapi.WithRouter`, then register documented routes through `openapi.UseRouter`.

```go
mux := httpx.New()
mux = openapi.WithRouter(mux, "User API", "1.0.0")

router := openapi.UseRouter(mux)
```

## Route Registration

Use:

- `router.Route(pattern, op.Options(...), handler)` to document and register an `httpx` error-returning handler together
- `router.Handle` or `router.HandleFunc` when you already have standard `net/http` handlers
- `router.Operation(pattern, op.Options(...))` when documentation and runtime registration happen separately

```go
router.Route("GET /users/{id}",
	op.Options(
		op.Summary("Get user"),
		op.PathParam[int]("id"),
		op.Response[User](200),
		op.Response[ErrorResponse](404),
	),
	func(w http.ResponseWriter, r *http.Request) error {
		return httpx.Send(w, User{ID: 1})
	},
)
```

## Schemas

Schemas are reflected from Go types automatically.

By default, reflection maps common Go shapes into OpenAPI schemas:

- strings to `type: string`
- booleans to `type: boolean`
- ints and uints to `type: integer`
- floats to `type: number`
- slices and arrays to array schemas
- maps to object schemas with `additionalProperties`
- structs to object schemas with reflected fields
- pointers to the underlying type with nullable behavior
- `time.Time` to `string` with `date-time` format

Struct reflection follows JSON tags. Unexported fields and `json:"-"` fields are skipped. Non-pointer fields are required by default unless tagged with `omitempty` or `omitzero`.

Prefer plain Go structs first. Add `schema.Schema()` methods or `schema.Field(...)` overrides when you want additional schema configuration beyond the default reflected shape.

For example, this struct works without extra config:

```go
type User struct {
	ID    int    `json:"id"`
	Email string `json:"email"`
}
```

Add schema config when you want more detail such as a reusable component reference or field validation:

```go
func (User) Schema() schema.Option {
	return schema.Options(
		schema.Reference("User"),
		schema.Field("email", schema.Email()),
	)
}
```

## Packages

The `openapi` tree in this repo currently includes:

- `openapi`: Use for the main router layer that binds OpenAPI generation to an `httpx` mux, registers documented routes, and serves JSON, YAML, or Scalar docs.
- `openapi/doc`: Use for top-level document metadata like version, summary, description, contact, license, servers, tags, security schemes, external docs, and webhooks.
- `openapi/op`: Use for operation-level metadata on a route such as summary, description, operation ID, tags, params, request body, responses, callbacks, per-operation security, and per-operation server overrides.
- `openapi/op/body`: Use when customizing a request body beyond the default `op.RequestBody[T]`, especially for descriptions, required/optional bodies, references, or extra content types.
- `openapi/op/callback`: Use when an operation documents outbound callbacks or webhook-style follow-up requests triggered by the server.
- `openapi/op/example`: Use when defining reusable named examples or adding summary, description, or external-value metadata to examples.
- `openapi/op/mediatype`: Use when customizing a concrete media type entry such as `application/json`, including examples, named examples, per-property encodings, or item schema details.
- `openapi/op/param`: Use when documenting individual path, query, header, or cookie parameters with descriptions, examples, required flags, serialization style, explode behavior, or references.
- `openapi/op/resp`: Use when customizing responses beyond the default `op.Response[T]`, especially for descriptions, headers, links, references, or extra content types.
- `openapi/op/resp/header`: Use when a response includes documented headers and those headers need schema, examples, required/deprecated flags, or serialization details.
- `openapi/op/resp/link`: Use when a response should describe a design-time link to another operation using runtime expressions, parameter mappings, request-body mappings, or server overrides.
- `openapi/pathitem`: Use when you want to build or reuse a whole OpenAPI path item, including shared path-level metadata, shared parameters, path-level servers, or multiple method handlers together.
- `openapi/schema`: Use when the reflected schema from a Go type needs overrides like titles, descriptions, formats, examples, enum values, validation constraints, references, object rules, or composition (`oneOf`, `allOf`, etc.). `schema.Type[T]()` and `schema.TypeOf(v, store)` are especially handy when composing schemas dynamically or passing type getters into helpers like `OneOf`, `AnyOf`, and `AllOf`.
- `openapi/securityscheme`: Use when defining reusable auth schemes in `components/securitySchemes`, including API key, HTTP auth, OAuth2 flows, OpenID Connect, bearer format, and related metadata.
- `openapi/server`: Use when a document, path, operation, or linked operation needs explicit server entries or templated server variables.
- `openapi/tag`: Use when defining top-level tag metadata such as descriptions, external docs, or OpenAPI 3.2 fields like summary, parent, and kind.

Useful schema entry points:

- `schema.New[T](store)` to reflect a type directly into a schema proxy
- `schema.Type[T]()` to produce a lazy type getter for composition helpers such as `schema.OneOf(...)`, `schema.AnyOf(...)`, and `schema.AllOf(...)`
- `schema.TypeOf(v, store)` when the schema should come from a dynamic runtime value instead of a static generic type

## Type-Driven Spec Hooks

Types can implement small interfaces to provide default OpenAPI behavior for different parts of the request or response spec.

Use these when you want the type itself to carry its own OpenAPI defaults instead of repeating options at each call site.

- `schema.Schema`: define schema defaults for the reflected Go type
- `param.Parameter`: define defaults for parameters using that type
- `body.RequestBody`: define defaults for request bodies using that type
- `mediatype.MediaType`: define defaults for media type entries using that type
- `resp.Response`: define defaults for responses using that type
- `header.Header`: define defaults for response headers using that type

Examples:

```go
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (User) Schema() schema.Option {
	return schema.Options(
		schema.Reference("User"),
		schema.Field("name", schema.MinLength(1)),
	)
}

func (User) Response() resp.Option {
	return resp.Options(
		resp.Reference("UserResponse"),
		resp.Description("User payload"),
	)
}

func (User) MediaType() mediatype.Option {
	return mediatype.Options(
		mediatype.NamedExample("default", User{ID: 1, Name: "Ada"}),
	)
}
```

A type may implement more than one of these interfaces at the same time. That is the intended pattern when one type should define its schema, examples, request-body defaults, and response defaults together.

## Reference Behavior

`Reference(...)` can be applied at multiple layers, and each layer stores that object into the matching OpenAPI components section:

For types that needs ref, the most reusable pattern is to put `Reference(...)` inside the type's own method.

```go
func (User) Schema() schema.Option {
	return schema.Reference("User")
}
```

Then call sites stay short:

```go
op.Response[User](200)
```

Inline references also work:

```go
op.Response[User](200, resp.Reference("UserResponse"))
```

Important rule: inline options are skipped only when the type itself sets `Reference(...)` through an implemented interface. This keeps the stored component stable and avoids mixing conflicting inline definitions with a type-owned reusable reference.

## Serving Docs

Use the router to expose generated docs:

```go
mux.Handle("GET /docs/openapi.json", router.OpenAPIJSONHandler())
mux.Handle("GET /docs/reference", router.ReferenceHandler(nil))
```
