package securityscheme

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// New creates a new security scheme of the given type.
// If a Reference option is set, the full scheme is stored in components/securitySchemes
// and a $ref object is returned instead.
func New(store *store.Store, name, schemeType string, opts ...Option) *v3.SecurityScheme {
	ss := &v3.SecurityScheme{Type: schemeType}
	for _, opt := range opts {
		opt(ss)
	}
	// If a reference is set, store the full security scheme in components
	// and return a security scheme $ref object.
	if store != nil {
		ss = &v3.SecurityScheme{Reference: store.SetSecurityScheme(name, ss)}
	}
	return ss
}
