package param

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	InPath   = "path"
	InQuery  = "query"
	InHeader = "header"
	InCookie = "cookie"
)

// Parameter may be implemented to set default parameter options on a type.
//
//	func (T) Parameter() param.Option
type Parameter interface {
	Parameter() Option
}

func New[T any](in, name string, store *store.Store, opts ...Option) *v3.Parameter {
	p := &v3.Parameter{
		Name:   name,
		In:     in,
		Schema: schema.New[T](store),
	}
	if t, ok := any(new(T)).(Parameter); ok {
		t.Parameter()(p, store)
	}
	if p.Reference == "" {
		// Skip inline options if reference is set on type
		for _, opt := range opts {
			opt(p, store)
		}
	}
	// If a reference is set, store the full parameter in components
	// and return a parameter $ref object.
	if store != nil && p.Reference != "" {
		p = store.SetParameter(p.Reference, p)
	}
	return p
}
