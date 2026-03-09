package tag

import (
	"github.com/pb33f/libopenapi/datamodel/high/base"
)

func New(name, description string, opts ...Option) *base.Tag {
	tag := &base.Tag{Name: name, Description: description}
	for _, opt := range opts {
		opt(tag)
	}
	return tag
}
