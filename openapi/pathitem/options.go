package pathitem

import (
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/op/param"
	"github.com/eriicafes/httpx/openapi/server"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

type Option func(*v3.PathItem, *store.Store)

// Reference sets a $ref to a named path item component in components/pathItems.
func Reference(name string) Option {
	return func(item *v3.PathItem, _ *store.Store) {
		item.Reference = name
	}
}

// Description sets the description for the path item.
func Description(s string) Option {
	return func(item *v3.PathItem, _ *store.Store) {
		item.Description = s
	}
}

// Summary sets a short summary for the path item.
func Summary(s string) Option {
	return func(item *v3.PathItem, _ *store.Store) {
		item.Summary = s
	}
}

// AdditionalOperation adds a named custom operation to the path item (OpenAPI 3.2+).
func AdditionalOperation(name string, opt op.Option) Option {
	return func(item *v3.PathItem, store *store.Store) {
		if item.AdditionalOperations == nil {
			item.AdditionalOperations = orderedmap.New[string, *v3.Operation]()
		}
		operation := &v3.Operation{}
		opt(operation, store)
		item.AdditionalOperations.Set(name, operation)
	}
}

// Server adds a server override for the path item.
func Server(url, description string, opts ...server.Option) Option {
	return func(item *v3.PathItem, _ *store.Store) {
		sv := &v3.Server{URL: url, Description: description}
		for _, opt := range opts {
			opt(sv)
		}
		item.Servers = append(item.Servers, sv)
	}
}

// Parameter adds a shared parameter to the path item.
// in must be one of "path", "query", "header", or "cookie".
func Parameter[T any](in, name string, opts ...param.Option) Option {
	return func(item *v3.PathItem, store *store.Store) {
		p := param.New[T](in, name, store, opts...)
		item.Parameters = append(item.Parameters, p)
	}
}
