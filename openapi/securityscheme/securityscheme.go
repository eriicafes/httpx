package securityscheme

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	TypeApiKey        = "apiKey"
	TypeHTTP          = "http"
	TypeOAuth2        = "oauth2"
	TypeOpenIdConnect = "openIdConnect"

	InHeader = "header"
	InQuery  = "query"
	InCookie = "cookie"
)

// New creates a new security scheme.
// The scheme is stored in components/securitySchemes under name.
func New(store *store.Store, name string, opts ...Option) *v3.SecurityScheme {
	ss := &v3.SecurityScheme{}
	for _, opt := range opts {
		opt(ss)
	}
	// If a reference is set, store the full security scheme in components
	// and return a security scheme $ref object.
	if store != nil {
		ss = store.SetSecurityScheme(name, ss)
	}
	return ss
}
