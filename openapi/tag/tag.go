package tag

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

// New creates a tag definition with the provided name, description, and options.
func New(name, description string, opts ...Option) *base.Tag {
	tag := &base.Tag{Name: name, Description: description}
	for _, opt := range opts {
		opt(tag)
	}
	return tag
}
