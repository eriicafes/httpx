package link

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

func New(store *store.Store, opts ...Option) *v3.Link {
	l := &v3.Link{}
	for _, opt := range opts {
		opt(l, store)
	}
	// If a reference is set, store the full link in components
	// and return a link $ref object.
	if store != nil && l.Reference != "" {
		l = store.SetLink(l.Reference, l)
	}
	return l
}
