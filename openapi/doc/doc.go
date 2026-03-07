package doc

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures an OpenAPI document.
type Option func(*v3.Document)

// OpenAPIVersion sets the OpenAPI specification version. Default: "3.1.0".
func OpenAPIVersion(version string) Option {
	return func(doc *v3.Document) {
		doc.Version = version
	}
}

// Info sets the API title and version.
func Info(title, version string) Option {
	return func(doc *v3.Document) {
		doc.Info.Title = title
		doc.Info.Version = version
	}
}

// Summary sets a short summary of the API.
func Summary(summary string) Option {
	return func(doc *v3.Document) {
		doc.Info.Summary = summary
	}
}

// Description sets the API description.
func Description(description string) Option {
	return func(doc *v3.Document) {
		doc.Info.Description = description
	}
}

// TermsOfService sets the URL to the terms of service.
func TermsOfService(url string) Option {
	return func(doc *v3.Document) {
		doc.Info.TermsOfService = url
	}
}

// Contact sets the contact information for the API.
func Contact(name, url, email string) Option {
	return func(doc *v3.Document) {
		doc.Info.Contact = &base.Contact{Name: name, URL: url, Email: email}
	}
}

// License sets the license for the API using a URL.
func License(name, url string) Option {
	return func(doc *v3.Document) {
		doc.Info.License = &base.License{Name: name, URL: url}
	}
}

// LicenseIdentifier sets the license using an SPDX identifier (OpenAPI 3.1).
func LicenseIdentifier(name, identifier string) Option {
	return func(doc *v3.Document) {
		doc.Info.License = &base.License{Name: name, Identifier: identifier}
	}
}

// ServerOption configures a server entry.
type ServerOption func(*v3.Server)

// ServerName sets the server name (OpenAPI 3.2+).
func ServerName(name string) ServerOption {
	return func(s *v3.Server) {
		s.Name = name
	}
}

// ServerVariable adds a URL template variable to the server.
func ServerVariable(name, defaultVal string, enum ...string) ServerOption {
	return func(s *v3.Server) {
		if s.Variables == nil {
			s.Variables = orderedmap.New[string, *v3.ServerVariable]()
		}
		s.Variables.Set(name, &v3.ServerVariable{Default: defaultVal, Enum: enum})
	}
}

// Server appends a server to the API document.
func Server(url, description string, opts ...ServerOption) Option {
	return func(doc *v3.Document) {
		sv := &v3.Server{URL: url, Description: description}
		for _, opt := range opts {
			opt(sv)
		}
		doc.Servers = append(doc.Servers, sv)
	}
}

// TagOption configures a tag entry.
type TagOption func(*base.Tag)

// TagSummary sets a short summary for the tag (OpenAPI 3.2+).
func TagSummary(summary string) TagOption {
	return func(t *base.Tag) {
		t.Summary = summary
	}
}

// TagExternalDocs attaches external documentation to the tag.
func TagExternalDocs(url, description string) TagOption {
	return func(t *base.Tag) {
		t.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Tag appends a tag definition to the API document.
func Tag(name, description string, opts ...TagOption) Option {
	return func(doc *v3.Document) {
		t := &base.Tag{Name: name, Description: description}
		for _, opt := range opts {
			opt(t)
		}
		doc.Tags = append(doc.Tags, t)
	}
}

// ExternalDocs sets the external documentation for the API.
func ExternalDocs(url, description string) Option {
	return func(doc *v3.Document) {
		doc.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Security appends a global security requirement.
// Pass an empty name to add an empty requirement (makes security optional globally).
func Security(name string, scopes ...string) Option {
	return func(doc *v3.Document) {
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
