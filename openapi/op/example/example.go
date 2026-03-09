package example

import (
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func New(store *store.Store, value any, opts ...Option) *base.Example {
	e := &base.Example{}
	if value != nil {
		e.Value = schema.ToYAMLNode(value)
	}
	for _, opt := range opts {
		opt(e)
	}
	// If a reference is set, store the full example in components
	// and return an example $ref object.
	if store != nil && e.Reference != "" {
		e = &base.Example{Reference: store.SetExample(e.Reference, e)}
	}
	return e
}
