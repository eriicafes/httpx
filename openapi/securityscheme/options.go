package securityscheme

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Option configures an OpenAPI security scheme.
type Option func(*v3.SecurityScheme)

// Reference sets a $ref to a named security scheme in components/securitySchemes.
func Reference(name string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Reference = name
	}
}

// Type sets the security scheme type.
func Type(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Type = s
	}
}

// ApiKey sets an "apiKey" security scheme type.
// in must be "header", "query", or "cookie".
func ApiKey(in, name string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Type = TypeApiKey
		ss.Name = name
		ss.In = in
	}
}

// HTTP sets an "http" security scheme type.
// scheme is the HTTP Authorization scheme name (e.g. "bearer", "basic").
func HTTP(scheme string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Type = TypeHTTP
		ss.Scheme = scheme
	}
}

// Description sets the security scheme description.
func Description(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Description = s
	}
}

// Name sets the name of the header, query, or cookie parameter for apiKey schemes.
func Name(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Name = s
	}
}

// In sets the location of the API key. Must be "header", "query", or "cookie".
func In(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.In = s
	}
}

// Scheme sets the HTTP Authorization scheme name (e.g. "bearer", "basic").
func Scheme(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.Scheme = s
	}
}

// BearerFormat sets a hint for the bearer token format, e.g. "JWT".
func BearerFormat(s string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.BearerFormat = s
	}
}

// OpenIdConnectUrl sets the OpenID Connect discovery URL.
func OpenIdConnectUrl(url string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.OpenIdConnectUrl = url
	}
}

// OAuth2MetadataUrl sets the OAuth2 metadata URL (OpenAPI 3.2+).
func OAuth2MetadataUrl(url string) Option {
	return func(ss *v3.SecurityScheme) {
		ss.OAuth2MetadataUrl = url
	}
}

// Deprecated marks the security scheme as deprecated (OpenAPI 3.2+).
func Deprecated() Option {
	return func(ss *v3.SecurityScheme) {
		ss.Deprecated = true
	}
}

// Flows sets the OAuth2 flows for the security scheme.
func Flows(opts ...OAuthFlowsOption) Option {
	return func(ss *v3.SecurityScheme) {
		if ss.Flows == nil {
			ss.Flows = &v3.OAuthFlows{}
		}
		for _, opt := range opts {
			opt(ss.Flows)
		}
	}
}
