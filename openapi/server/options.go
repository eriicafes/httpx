package server

import (
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Option configures a server entry.
type Option func(*v3.Server)

// Name sets the server name (OpenAPI 3.2+).
func Name(name string) Option {
	return func(s *v3.Server) {
		s.Name = name
	}
}

// Variable adds a URL template variable to the server.
func Variable(name, defaultVal string, opts ...VariableOption) Option {
	return func(s *v3.Server) {
		if s.Variables == nil {
			s.Variables = orderedmap.New[string, *v3.ServerVariable]()
		}
		sv := &v3.ServerVariable{Default: defaultVal}
		for _, opt := range opts {
			opt(sv)
		}
		s.Variables.Set(name, sv)
	}
}
