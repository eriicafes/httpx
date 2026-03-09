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

// GetSchema returns a stored schema for t.
func (s *Store) GetSchema(t reflect.Type) (*base.SchemaProxy, bool) {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	if ref, ok := s.types[t]; ok {
		return base.CreateSchemaProxyRef(ref), true
	}
	return nil, false
}

// SetSchema stores a schema for t in components/schemas if not already stored,
// and returns the full $ref string.
func (s *Store) SetSchema(t reflect.Type, name string, proxy *base.SchemaProxy) string {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	ref := "#/components/schemas/" + name
	if _, ok := s.types[t]; !ok {
		s.types[t] = ref
		if s.components.Schemas == nil {
			s.components.Schemas = orderedmap.New[string, *base.SchemaProxy]()
		}
		s.components.Schemas.Set(name, proxy)
	}
	return ref
}

// --- Parameter ---

// GetParameter returns a stored parameter by name.
func (s *Store) GetParameter(name string) (*v3.Parameter, bool) {
	return s.components.Parameters.Get(name)
}

// SetParameter stores a parameter in components/parameters if not already stored,
// and returns the full $ref string.
func (s *Store) SetParameter(name string, p *v3.Parameter) string {
	ref := "#/components/parameters/" + name
	if s.components.Parameters == nil {
		s.components.Parameters = orderedmap.New[string, *v3.Parameter]()
	}
	if _, ok := s.components.Parameters.Get(name); !ok {
		s.components.Parameters.Set(name, p)
	}
	return ref
}

// --- Request Body ---

// GetRequestBody returns a stored request body by name.
func (s *Store) GetRequestBody(name string) (*v3.RequestBody, bool) {
	return s.components.RequestBodies.Get(name)
}

// SetRequestBody stores a request body in components/requestBodies if not already stored,
// and returns the full $ref string.
func (s *Store) SetRequestBody(name string, b *v3.RequestBody) string {
	ref := "#/components/requestBodies/" + name
	if s.components.RequestBodies == nil {
		s.components.RequestBodies = orderedmap.New[string, *v3.RequestBody]()
	}
	if _, ok := s.components.RequestBodies.Get(name); !ok {
		s.components.RequestBodies.Set(name, b)
	}
	return ref
}

// --- Response ---

// GetResponse returns a stored response by name.
func (s *Store) GetResponse(name string) (*v3.Response, bool) {
	return s.components.Responses.Get(name)
}

// SetResponse stores a response in components/responses if not already stored,
// and returns the full $ref string.
func (s *Store) SetResponse(name string, r *v3.Response) string {
	ref := "#/components/responses/" + name
	if s.components.Responses == nil {
		s.components.Responses = orderedmap.New[string, *v3.Response]()
	}
	if _, ok := s.components.Responses.Get(name); !ok {
		s.components.Responses.Set(name, r)
	}
	return ref
}

// --- Header ---

// GetHeader returns a stored header by name.
func (s *Store) GetHeader(name string) (*v3.Header, bool) {
	return s.components.Headers.Get(name)
}

// SetHeader stores a header in components/headers if not already stored,
// and returns the full $ref string.
func (s *Store) SetHeader(name string, h *v3.Header) string {
	ref := "#/components/headers/" + name
	if s.components.Headers == nil {
		s.components.Headers = orderedmap.New[string, *v3.Header]()
	}
	if _, ok := s.components.Headers.Get(name); !ok {
		s.components.Headers.Set(name, h)
	}
	return ref
}

// --- Path Item ---

// GetPathItem returns a stored path item by name.
func (s *Store) GetPathItem(name string) (*v3.PathItem, bool) {
	return s.components.PathItems.Get(name)
}

// SetPathItem stores a path item in components/pathItems if not already stored,
// and returns the full $ref string.
func (s *Store) SetPathItem(name string, h *v3.PathItem) string {
	ref := "#/components/pathItems/" + name
	if s.components.PathItems == nil {
		s.components.PathItems = orderedmap.New[string, *v3.PathItem]()
	}
	if _, ok := s.components.PathItems.Get(name); !ok {
		s.components.PathItems.Set(name, h)
	}
	return ref
}

// --- Link ---

// GetLink returns a stored link by name.
func (s *Store) GetLink(name string) (*v3.Link, bool) {
	return s.components.Links.Get(name)
}

// SetLink stores a link in components/links if not already stored,
// and returns the full $ref string.
func (s *Store) SetLink(name string, l *v3.Link) string {
	ref := "#/components/links/" + name
	if s.components.Links == nil {
		s.components.Links = orderedmap.New[string, *v3.Link]()
	}
	if _, ok := s.components.Links.Get(name); !ok {
		s.components.Links.Set(name, l)
	}
	return ref
}

// --- Example ---

// GetExample returns a stored example by name.
func (s *Store) GetExample(name string) (*base.Example, bool) {
	return s.components.Examples.Get(name)
}

// SetExample stores an example in components/examples if not already stored,
// and returns the full $ref string.
func (s *Store) SetExample(name string, e *base.Example) string {
	ref := "#/components/examples/" + name
	if s.components.Examples == nil {
		s.components.Examples = orderedmap.New[string, *base.Example]()
	}
	if _, ok := s.components.Examples.Get(name); !ok {
		s.components.Examples.Set(name, e)
	}
	return ref
}

// --- Callback ---

// GetCallback returns a stored callback by name.
func (s *Store) GetCallback(name string) (*v3.Callback, bool) {
	return s.components.Callbacks.Get(name)
}

// SetCallback stores a callback in components/callbacks if not already stored,
// and returns the full $ref string.
func (s *Store) SetCallback(name string, cb *v3.Callback) string {
	ref := "#/components/callbacks/" + name
	if s.components.Callbacks == nil {
		s.components.Callbacks = orderedmap.New[string, *v3.Callback]()
	}
	if _, ok := s.components.Callbacks.Get(name); !ok {
		s.components.Callbacks.Set(name, cb)
	}
	return ref
}

// --- Security Scheme ---

// GetSecurityScheme returns a stored security scheme by name.
func (s *Store) GetSecurityScheme(name string) (*v3.SecurityScheme, bool) {
	return s.components.SecuritySchemes.Get(name)
}

// SetSecurityScheme stores a security scheme in components/securitySchemes if not already stored,
// and returns the full $ref string.
func (s *Store) SetSecurityScheme(name string, ss *v3.SecurityScheme) string {
	ref := "#/components/securitySchemes/" + name
	if s.components.SecuritySchemes == nil {
		s.components.SecuritySchemes = orderedmap.New[string, *v3.SecurityScheme]()
	}
	if _, ok := s.components.SecuritySchemes.Get(name); !ok {
		s.components.SecuritySchemes.Set(name, ss)
	}
	return ref
}
