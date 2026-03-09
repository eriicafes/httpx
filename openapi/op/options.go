package op

import (
	"net/http"
	"strconv"

	"github.com/eriicafes/httpx/openapi/op/body"
	"github.com/eriicafes/httpx/openapi/op/callback"
	"github.com/eriicafes/httpx/openapi/op/param"
	"github.com/eriicafes/httpx/openapi/op/resp"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// NoContent is used as the type parameter for RequestBody or Response when there is no content body.
// It implements the noContent marker interface checked inline by body.New and resp.New.
type NoContent struct{}

func (NoContent) NoContent() {}

// Option configures an OpenAPI operation.
type Option func(*v3.Operation, *store.Store)

// Options combines multiple options into one.
func Options(opts ...Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		for _, opt := range opts {
			opt(operation, store)
		}
	}
}

// Summary sets a short summary of the operation.
func Summary(s string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.Summary = s
	}
}

// Description sets the operation description.
func Description(s string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.Description = s
	}
}

// OperationId sets the unique operation identifier.
func OperationId(id string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.OperationId = id
	}
}

// Tags applies tags to the operation.
func Tags(tags ...string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.Tags = append(op.Tags, tags...)
	}
}

// Deprecated marks the operation as deprecated. Default: not deprecated.
func Deprecated() Option {
	deprecated := true
	return func(op *v3.Operation, _ *store.Store) {
		op.Deprecated = &deprecated
	}
}

// ExternalDocs sets an external documentation link for the operation.
func ExternalDocs(url, description string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Server adds a server override for this specific operation.
func Server(url, description string) Option {
	return func(op *v3.Operation, _ *store.Store) {
		op.Servers = append(op.Servers, &v3.Server{URL: url, Description: description})
	}
}

// Security appends a security requirement to the operation.
// Pass an empty name to add an empty requirement (makes security optional for this operation).
func Security(name string, scopes ...string) Option {
	return func(op *v3.Operation, _ *store.Store) {
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
	return func(operation *v3.Operation, store *store.Store) {
		opts = append([]param.Option{param.Required(true)}, opts...)
		p := param.New[T]("path", name, store, opts...)
		operation.Parameters = append(operation.Parameters, p)
	}
}

// QueryParam adds an optional query parameter to the operation.
func QueryParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		p := param.New[T]("query", name, store, opts...)
		operation.Parameters = append(operation.Parameters, p)
	}
}

// HeaderParam adds an optional header parameter to the operation.
func HeaderParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		p := param.New[T]("header", name, store, opts...)
		operation.Parameters = append(operation.Parameters, p)
	}
}

// CookieParam adds an optional cookie parameter to the operation.
func CookieParam[T any](name string, opts ...param.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		p := param.New[T]("cookie", name, store, opts...)
		operation.Parameters = append(operation.Parameters, p)
	}
}

// RequestBody sets the request body for the operation.
func RequestBody[T any](opts ...body.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		operation.RequestBody = body.New[T](store, opts...)
	}
}

// Response sets the response for the given HTTP status code.
// The description defaults to the HTTP status text.
func Response[T any](status int, opts ...resp.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		if operation.Responses == nil {
			operation.Responses = &v3.Responses{}
		}
		if operation.Responses.Codes == nil {
			operation.Responses.Codes = orderedmap.New[string, *v3.Response]()
		}
		r := resp.New[T](store, http.StatusText(status), opts...)
		operation.Responses.Codes.Set(strconv.Itoa(status), r)
	}
}

// Callback adds a named callback to the operation.
func Callback(name string, opts ...callback.Option) Option {
	return func(operation *v3.Operation, store *store.Store) {
		if operation.Callbacks == nil {
			operation.Callbacks = orderedmap.New[string, *v3.Callback]()
		}
		operation.Callbacks.Set(name, callback.New(store, opts...))
	}
}
