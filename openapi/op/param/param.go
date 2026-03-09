package param

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Parameter is implemented by types that provide default parameter options.
type Parameter interface {
	Parameter() Option
}

func New[T any](in, name string, store *store.Store, opts ...Option) *v3.Parameter {
	p := &v3.Parameter{
		Name:   name,
		In:     in,
		Schema: schema.New[T](store),
	}
	if pm, ok := any(new(T)).(Parameter); ok {
		pm.Parameter()(p, store)
	}
	for _, opt := range opts {
		opt(p, store)
	}
	// If a reference is set, store the full parameter in components
	// and return a parameter $ref object.
	if store != nil && p.Reference != "" {
		p = &v3.Parameter{Reference: store.SetParameter(p.Reference, p)}
	}
	return p
}
