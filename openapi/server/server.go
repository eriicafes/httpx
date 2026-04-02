package server

import v3 "github.com/pb33f/libopenapi/datamodel/high/v3"

// New creates a server entry with the provided URL, description, and options.
func New(url, description string, opts ...Option) *v3.Server {
	sv := &v3.Server{URL: url, Description: description}
	for _, opt := range opts {
		opt(sv)
	}
	return sv
}
