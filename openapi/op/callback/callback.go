package callback

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// New builds a callback and stores it as a component when referenced.
func New(store *store.Store, opts ...Option) *v3.Callback {
	cb := &v3.Callback{
		Expression: orderedmap.New[string, *v3.PathItem](),
	}
	for _, opt := range opts {
		opt(cb, store)
	}
	// If a reference is set, store the full callback in components
	// and return a callback $ref object.
	if store != nil && cb.Reference != "" {
		cb = store.SetCallback(cb.Reference, cb)
	}
	return cb
}
