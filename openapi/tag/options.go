package tag

import "github.com/pb33f/libopenapi/datamodel/high/base"

// Option configures a tag entry.
type Option func(*base.Tag)

// Summary sets a short summary for the tag (OpenAPI 3.2+).
func Summary(summary string) Option {
	return func(t *base.Tag) {
		t.Summary = summary
	}
}

// ExternalDocs attaches external documentation to the tag.
func ExternalDocs(url, description string) Option {
	return func(t *base.Tag) {
		t.ExternalDocs = &base.ExternalDoc{URL: url, Description: description}
	}
}

// Parent sets the parent tag name (OpenAPI 3.2+).
func Parent(parent string) Option {
	return func(t *base.Tag) {
		t.Parent = parent
	}
}

// Kind sets the tag kind (OpenAPI 3.2+).
func Kind(kind string) Option {
	return func(t *base.Tag) {
		t.Kind = kind
	}
}
