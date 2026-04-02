package securityscheme

import (
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

const (
	// TypeApiKey identifies an apiKey security scheme.
	TypeApiKey = "apiKey"
	// TypeHTTP identifies an http security scheme.
	TypeHTTP = "http"
	// TypeOAuth2 identifies an oauth2 security scheme.
	TypeOAuth2 = "oauth2"
	// TypeOpenIdConnect identifies an openIdConnect security scheme.
	TypeOpenIdConnect = "openIdConnect"

	// InHeader identifies a header API key location.
	InHeader = "header"
	// InQuery identifies a query API key location.
	InQuery = "query"
	// InCookie identifies a cookie API key location.
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
