package store

import (
	"reflect"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Store tracks component references by writing them into the OpenAPI document's
// components object.
type Store struct {
	types      map[reflect.Type]string
	components *v3.Components
}

// New creates a [Store] backed by the given components object.
func New(components *v3.Components) *Store {
	return &Store{
		types:      make(map[reflect.Type]string),
		components: components,
	}
}

// --- Schema ---

// GetSchema returns a stored schema $ref for t.
func (s *Store) GetSchema(t reflect.Type) (*base.SchemaProxy, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if name, ok := s.types[t]; ok {
		return base.CreateSchemaProxyRef("#/components/schemas/" + name), true
	}
	return nil, false
}

// SetSchema stores a schema for t in components/schemas if not already stored,
// and returns the $ref.
func (s *Store) SetSchema(t reflect.Type, name string, p *base.SchemaProxy) *base.SchemaProxy {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if _, ok := s.types[t]; !ok {
		s.types[t] = name
		if s.components.Schemas == nil {
			s.components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
		}
		s.components.Schemas.Set(name, p)
	}
	return base.CreateSchemaProxyRef("#/components/schemas/" + name)
}

// --- Parameter ---

// GetParameter returns a stored parameter $ref.
func (s *Store) GetParameter(name string) (*v3.Parameter, bool) {
	if s.components.Parameters == nil {
		return nil, false
	}
	if _, ok := s.components.Parameters.Get(name); !ok {
		return nil, false
	}

	return v3.CreateParameterRef("#/components/parameters/" + name), true
}

// SetParameter stores a parameter in components/parameters if not already stored,
// and returns the $ref.
func (s *Store) SetParameter(name string, p *v3.Parameter) *v3.Parameter {
	if s.components.Parameters == nil {
		s.components.Parameters = orderedmap.New[string, *v3.Parameter]()
	}
	if _, ok := s.components.Parameters.Get(name); !ok {
		p.Reference = "" // unset the reference field
		s.components.Parameters.Set(name, p)
	}
	return v3.CreateParameterRef("#/components/parameters/" + name)
}

// --- Request Body ---

// GetRequestBody returns a stored request body $ref.
func (s *Store) GetRequestBody(name string) (*v3.RequestBody, bool) {
	if s.components.RequestBodies == nil {
		return nil, false
	}
	if _, ok := s.components.RequestBodies.Get(name); !ok {
		return nil, false
	}
	return v3.CreateRequestBodyRef("#/components/requestBodies/" + name), true
}

// SetRequestBody stores a request body in components/requestBodies if not already stored,
// and returns the $ref.
func (s *Store) SetRequestBody(name string, b *v3.RequestBody) *v3.RequestBody {
	if s.components.RequestBodies == nil {
		s.components.RequestBodies = orderedmap.New[string, *v3.RequestBody]()
	}
	if _, ok := s.components.RequestBodies.Get(name); !ok {
		b.Reference = "" // unset the reference field
		s.components.RequestBodies.Set(name, b)
	}
	return v3.CreateRequestBodyRef("#/components/requestBodies/" + name)
}

// --- Response ---

// GetResponse returns a stored response $ref.
func (s *Store) GetResponse(name string) (*v3.Response, bool) {
	if s.components.Responses == nil {
		return nil, false
	}
	if _, ok := s.components.Responses.Get(name); !ok {
		return nil, false
	}
	return v3.CreateResponseRef("#/components/responses/" + name), true
}

// SetResponse stores a response in components/responses if not already stored,
// and returns the $ref.
func (s *Store) SetResponse(name string, r *v3.Response) *v3.Response {
	if s.components.Responses == nil {
		s.components.Responses = orderedmap.New[string, *v3.Response]()
	}
	if _, ok := s.components.Responses.Get(name); !ok {
		r.Reference = "" // unset the reference field
		s.components.Responses.Set(name, r)
	}
	return v3.CreateResponseRef("#/components/responses/" + name)
}

// --- Header ---

// GetHeader returns a stored header $ref.
func (s *Store) GetHeader(name string) (*v3.Header, bool) {
	if s.components.Headers == nil {
		return nil, false
	}
	if _, ok := s.components.Headers.Get(name); !ok {
		return nil, false
	}
	return v3.CreateHeaderRef("#/components/headers/" + name), true
}

// SetHeader stores a header in components/headers if not already stored,
// and returns the $ref.
func (s *Store) SetHeader(name string, h *v3.Header) *v3.Header {
	if s.components.Headers == nil {
		s.components.Headers = orderedmap.New[string, *v3.Header]()
	}
	if _, ok := s.components.Headers.Get(name); !ok {
		h.Reference = "" // unset the reference field
		s.components.Headers.Set(name, h)
	}
	return v3.CreateHeaderRef("#/components/headers/" + name)
}

// --- Path Item ---

// GetPathItem returns a stored path item $ref.
func (s *Store) GetPathItem(name string) (*v3.PathItem, bool) {
	if s.components.PathItems == nil {
		return nil, false
	}
	if _, ok := s.components.PathItems.Get(name); !ok {
		return nil, false
	}
	return v3.CreatePathItemRef("#/components/pathItems/" + name), true
}

// SetPathItem stores a path item in components/pathItems if not already stored,
// and returns the $ref.
func (s *Store) SetPathItem(name string, p *v3.PathItem) *v3.PathItem {
	if s.components.PathItems == nil {
		s.components.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if _, ok := s.components.PathItems.Get(name); !ok {
		p.Reference = "" // unset the reference field
		s.components.PathItems.Set(name, p)
	}
	return v3.CreatePathItemRef("#/components/pathItems/" + name)
}

// --- Link ---

// GetLink returns a stored link $ref.
func (s *Store) GetLink(name string) (*v3.Link, bool) {
	if s.components.Links == nil {
		return nil, false
	}
	if _, ok := s.components.Links.Get(name); !ok {
		return nil, false
	}
	return v3.CreateLinkRef("#/components/links/" + name), true
}

// SetLink stores a link in components/links if not already stored,
// and returns the $ref.
func (s *Store) SetLink(name string, l *v3.Link) *v3.Link {
	if s.components.Links == nil {
		s.components.Links = orderedmap.New[string, *v3.Link]()
	}
	if _, ok := s.components.Links.Get(name); !ok {
		l.Reference = "" // unset the reference field
		s.components.Links.Set(name, l)
	}
	return v3.CreateLinkRef("#/components/links/" + name)
}

// --- Example ---

// GetExample returns a stored example $ref.
func (s *Store) GetExample(name string) (*base.Example, bool) {
	if s.components.Examples == nil {
		return nil, false
	}
	if _, ok := s.components.Examples.Get(name); !ok {
		return nil, false
	}
	return base.CreateExampleRef("#/components/examples/" + name), true
}

// SetExample stores an example in components/examples if not already stored,
// and returns the $ref.
func (s *Store) SetExample(name string, e *base.Example) *base.Example {
	if s.components.Examples == nil {
		s.components.Examples = orderedmap.New[string, *base.Example]()
	}
	if _, ok := s.components.Examples.Get(name); !ok {
		e.Reference = "" // unset the reference field
		s.components.Examples.Set(name, e)
	}
	return base.CreateExampleRef("#/components/examples/" + name)
}

// --- Callback ---

// GetCallback returns a stored callback $ref.
func (s *Store) GetCallback(name string) (*v3.Callback, bool) {
	if s.components.Callbacks == nil {
		return nil, false
	}
	if _, ok := s.components.Callbacks.Get(name); !ok {
		return nil, false
	}
	return v3.CreateCallbackRef("#/components/callbacks/" + name), true
}

// SetCallback stores a callback in components/callbacks if not already stored,
// and returns the $ref.
func (s *Store) SetCallback(name string, cb *v3.Callback) *v3.Callback {
	if s.components.Callbacks == nil {
		s.components.Callbacks = orderedmap.New[string, *v3.Callback]()
	}
	if _, ok := s.components.Callbacks.Get(name); !ok {
		cb.Reference = "" // unset the reference field
		s.components.Callbacks.Set(name, cb)
	}
	return v3.CreateCallbackRef("#/components/callbacks/" + name)
}

// --- Security Scheme ---

// GetSecurityScheme returns a stored security scheme $ref.
func (s *Store) GetSecurityScheme(name string) (*v3.SecurityScheme, bool) {
	if s.components.SecuritySchemes == nil {
		return nil, false
	}
	if _, ok := s.components.SecuritySchemes.Get(name); !ok {
		return nil, false
	}
	return v3.CreateSecuritySchemeRef("#/components/securitySchemes/" + name), true
}

// SetSecurityScheme stores a security scheme in components/securitySchemes if not already stored,
// and returns the $ref.
func (s *Store) SetSecurityScheme(name string, ss *v3.SecurityScheme) *v3.SecurityScheme {
	if s.components.SecuritySchemes == nil {
		s.components.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}
	if _, ok := s.components.SecuritySchemes.Get(name); !ok {
		ss.Reference = "" // unset the reference field
		s.components.SecuritySchemes.Set(name, ss)
	}
	return v3.CreateSecuritySchemeRef("#/components/securitySchemes/" + name)
}
