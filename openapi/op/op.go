package op

import (
	"net/http"
	"reflect"
	"strconv"

	"github.com/eriicafes/httpx/openapi/body"
	"github.com/eriicafes/httpx/openapi/param"
	"github.com/eriicafes/httpx/openapi/resp"
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI operation.
type Option func(*v3.Operation, *schema.Registry)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		for _, opt := range opts {
			opt(operation, registry)
		}
	}
}

// Summary sets a short summary of the operation.
func Summary(s string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.Summary = s
	}
}

// Description sets the operation description.
func Description(s string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.Description = s
	}
}

// OperationId sets the unique operation identifier.
func OperationId(id string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.OperationId = id
	}
}

// Tags applies tags to the operation.
func Tags(tags ...string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.Tags = append(op.Tags, tags...)
	}
}

// Deprecated marks the operation as deprecated. Default: not deprecated.
func Deprecated() Option {
	deprecated := true
	return func(op *v3.Operation, _ *schema.Registry) {
		op.Deprecated = &deprecated
	}
}

// ExternalDocs sets an external documentation link for the operation.
func ExternalDocs(url, description string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Server adds a server override for this specific operation.
func Server(url, description string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		op.Servers = append(op.Servers, &v3.Server{URL: url, Description: description})
	}
}

// Security appends a security requirement to the operation.
// Pass an empty name to add an empty requirement (makes security optional for this operation).
func Security(name string, scopes ...string) Option {
	return func(op *v3.Operation, _ *schema.Registry) {
		req := &base.SecurityRequirement{
			Requirements: orderedmap.New[string, []string](),
		}
		if name != "" {
			req.Requirements.Set(name, scopes)
		} else {
			req.ContainsEmptyRequirement = true
		}
		op.Security = append(op.Security, req)
	}
}

// PathParam adds a required path parameter to the operation.
func PathParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		required := true
		p := &v3.Parameter{
			Name:     name,
			In:       "path",
			Schema:   schema.ReflectType(reflect.TypeFor[T](), registry),
			Required: &required,
		}
		for _, opt := range opts {
			opt(p)
		}
		operation.Parameters = append(operation.Parameters, p)
	}
}

// QueryParam adds an optional query parameter to the operation.
func QueryParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		p := &v3.Parameter{
			Name:   name,
			In:     "query",
			Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
		}
		for _, opt := range opts {
			opt(p)
		}
		operation.Parameters = append(operation.Parameters, p)
	}
}

// HeaderParam adds an optional header parameter to the operation.
func HeaderParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		p := &v3.Parameter{
			Name:   name,
			In:     "header",
			Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
		}
		for _, opt := range opts {
			opt(p)
		}
		operation.Parameters = append(operation.Parameters, p)
	}
}

// CookieParam adds an optional cookie parameter to the operation.
func CookieParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		p := &v3.Parameter{
			Name:   name,
			In:     "cookie",
			Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
		}
		for _, opt := range opts {
			opt(p)
		}
		operation.Parameters = append(operation.Parameters, p)
	}
}

// RequestBody sets the request body for the operation.
func RequestBody[T any](opts ...body.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		required := true
		b := &v3.RequestBody{
			Required: &required,
			Content: orderedmap.FromPairs(
				orderedmap.NewPair("application/json", &v3.MediaType{
					Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
				}),
			),
		}
		for _, opt := range opts {
			opt(b, registry)
		}
		operation.RequestBody = b
	}
}

// Response sets the response for the given HTTP status code.
// The description defaults to the HTTP status text.
func Response[T any](status int, opts ...resp.Option) Option {
	return func(operation *v3.Operation, registry *schema.Registry) {
		if operation.Responses == nil {
			operation.Responses = &v3.Responses{
				Codes: orderedmap.New[string, *v3.Response](),
			}
		}
		r := &v3.Response{Description: http.StatusText(status)}
		var zero T
		if _, ok := any(zero).(NoContent); !ok {
			r.Content = orderedmap.FromPairs(
				orderedmap.NewPair("application/json", &v3.MediaType{
					Schema: schema.ReflectType(reflect.TypeFor[T](), registry),
				}),
			)
		}
		for _, opt := range opts {
			opt(r, registry)
		}
		operation.Responses.Codes.Set(strconv.Itoa(status), r)
	}
}

// NoContent is used as the type parameter for Response when the response has no body.
type NoContent struct{}
