package doc

import (
	"github.com/eriicafes/httpx/openapi/pathitem"
	"github.com/eriicafes/httpx/openapi/securityscheme"
	"github.com/eriicafes/httpx/openapi/server"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/eriicafes/httpx/openapi/tag"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI document.
type Option func(*v3.Document, *store.Store)

// OpenAPIVersion sets the OpenAPI specification version. Default: "3.1.0".
func OpenAPIVersion(version string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Version = version
	}
}

// Info sets the API title and version.
func Info(title, version string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.Title = title
		doc.Info.Version = version
	}
}

// Summary sets a short summary of the API.
func Summary(summary string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.Summary = summary
	}
}

// Description sets the API description.
func Description(description string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.Description = description
	}
}

// TermsOfService sets the URL to the terms of service.
func TermsOfService(url string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.TermsOfService = url
	}
}

// Contact sets the contact information for the API.
func Contact(name, url, email string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.Contact = &base.Contact{Name: name, URL: url, Email: email}
	}
}

// License sets the license for the API.
// Provide url for a link to the full license text, or identifier for an SPDX identifier (OpenAPI 3.1).
// Only one of url or identifier should be set; if both are provided, url takes precedence per the spec.
func License(name, url, identifier string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.Info.License = &base.License{Name: name, URL: url, Identifier: identifier}
	}
}

// Server appends a server to the API document.
func Server(url, description string, opts ...server.Option) Option {
	return func(doc *v3.Document, _ *store.Store) {
		sv := server.New(url, description, opts...)
		doc.Servers = append(doc.Servers, sv)
	}
}

// Security appends a global security requirement.
// Pass an empty name to add an empty requirement (makes security optional globally).
func Security(name string, scopes ...string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		req := &base.SecurityRequirement{
			Requirements: orderedmap.New[string, []string](),
		}
		if name != "" {
			req.Requirements.Set(name, scopes)
		} else {
			req.ContainsEmptyRequirement = true
		}
		doc.Security = append(doc.Security, req)
	}
}

// SecurityScheme registers a named security scheme in components/securitySchemes.
func SecurityScheme(name, schemeType string, opts ...securityscheme.Option) Option {
	return func(doc *v3.Document, store *store.Store) {
		securityscheme.New(store, name, schemeType, opts...)
	}
}

// Tag appends a tag definition to the API document.
func Tag(name, description string, opts ...tag.Option) Option {
	return func(doc *v3.Document, _ *store.Store) {
		t := tag.New(name, description, opts...)
		doc.Tags = append(doc.Tags, t)
	}
}

// ExternalDocs sets the external documentation for the API.
func ExternalDocs(url, description string) Option {
	return func(doc *v3.Document, _ *store.Store) {
		doc.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Webhook adds a named webhook path to the API document.
func Webhook(name string, p *pathitem.Path) Option {
	return func(d *v3.Document, store *store.Store) {
		if d.Webhooks == nil {
			d.Webhooks = orderedmap.New[string, *v3.PathItem]()
		}
		d.Webhooks.Set(name, p.Item(store))
	}
}
