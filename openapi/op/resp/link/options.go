package link

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI response link.
type Option func(*v3.Link, *store.Store)

// Reference sets a $ref to a named link component in components/links.
func Reference(name string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.Reference = name
	}
}

// OperationRef sets a relative or absolute URI reference to an OAS operation.
func OperationRef(ref string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.OperationRef = ref
	}
}

// OperationId sets the name of an existing resolvable OAS operation.
func OperationId(id string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.OperationId = id
	}
}

// Parameter adds a parameter name to runtime expression mapping for the linked operation.
func Parameter(name, expression string) Option {
	return func(l *v3.Link, _ *store.Store) {
		if l.Parameters == nil {
			l.Parameters = orderedmap.New[string, string]()
		}
		l.Parameters.Set(name, expression)
	}
}

// RequestBody sets a runtime expression for the request body of the linked operation.
func RequestBody(expression string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.RequestBody = expression
	}
}

// Description sets the link description.
func Description(s string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.Description = s
	}
}

// Server sets a server to use for the linked operation.
func Server(url, description string) Option {
	return func(l *v3.Link, _ *store.Store) {
		l.Server = &v3.Server{URL: url, Description: description}
	}
}
