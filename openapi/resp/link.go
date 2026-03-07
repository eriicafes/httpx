package resp

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// LinkOption configures an OpenAPI response link.
type LinkOption func(*v3.Link)

// LinkOperationRef sets a relative or absolute URI reference to an OAS operation.
func LinkOperationRef(ref string) LinkOption {
	return func(l *v3.Link) {
		l.OperationRef = ref
	}
}

// LinkOperationId sets the name of an existing resolvable OAS operation.
func LinkOperationId(id string) LinkOption {
	return func(l *v3.Link) {
		l.OperationId = id
	}
}

// LinkParameter adds a parameter name to runtime expression mapping for the linked operation.
func LinkParameter(name, expression string) LinkOption {
	return func(l *v3.Link) {
		if l.Parameters == nil {
			l.Parameters = orderedmap.New[string, string]()
		}
		l.Parameters.Set(name, expression)
	}
}

// LinkRequestBody sets a runtime expression for the request body of the linked operation.
func LinkRequestBody(expression string) LinkOption {
	return func(l *v3.Link) {
		l.RequestBody = expression
	}
}

// LinkDescription sets the link description.
func LinkDescription(s string) LinkOption {
	return func(l *v3.Link) {
		l.Description = s
	}
}

// LinkServer sets a server to use for the linked operation.
func LinkServer(url, description string) LinkOption {
	return func(l *v3.Link) {
		l.Server = &v3.Server{URL: url, Description: description}
	}
}

// LinkReference sets a $ref to an external link definition.
func LinkReference(ref string) LinkOption {
	return func(l *v3.Link) {
		l.Reference = ref
	}
}
