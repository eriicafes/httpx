package pathitem

import (
	"net/http"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// Path collects path item options, operations, and optional handlers for a single path.
type Path struct {
	opts       []Option
	operations map[string]op.Option
	handlers   map[string]httpx.HandlerFunc
}

// New creates a Path that collects path item options, operations, and handlers.
func New(opts ...Option) *Path {
	return &Path{
		opts:       opts,
		operations: make(map[string]op.Option),
		handlers:   make(map[string]httpx.HandlerFunc),
	}
}

// Operation registers an OpenAPI operation for the given method.
// If no method is present the operation is registered for all supported methods.
// Use this when the handler is registered separately.
func (p *Path) Operation(method string, opt op.Option) {
	p.operations[method] = opt
}

// Route registers an OpenAPI operation and a handler on the underlying mux.
func (p *Path) Route(method string, opt op.Option, handler func(http.ResponseWriter, *http.Request) error) {
	p.Operation(method, opt)
	p.handlers[method] = handler
}

// Handle registers an OpenAPI operation and a handler on the underlying mux.
func (p *Path) Handle(method string, opt op.Option, handler http.Handler) {
	p.Operation(method, opt)
	p.handlers[method] = func(w http.ResponseWriter, r *http.Request) error {
		handler.ServeHTTP(w, r)
		return nil
	}
}

// HandleFunc registers an OpenAPI operation and a handler func on the underlying mux.
func (p *Path) HandleFunc(method string, opt op.Option, handler func(http.ResponseWriter, *http.Request)) {
	p.Operation(method, opt)
	p.handlers[method] = func(w http.ResponseWriter, r *http.Request) error {
		handler(w, r)
		return nil
	}
}

// PathItem returns a builder function that constructs the path item from p.
func (p *Path) PathItem() func(*store.Store) *v3.PathItem {
	return func(store *store.Store) *v3.PathItem {
		return GetPathItem(p, store, nil)
	}
}

// GetPathItem constructs an OpenAPI path item from p's options and operations.
// If a reference is set, the full item is stored in components and a $ref object is returned.
func GetPathItem(p *Path, store *store.Store, item *v3.PathItem) *v3.PathItem {
	if item == nil {
		item = &v3.PathItem{}
	}
	for _, opt := range p.opts {
		opt(item, store)
	}
	// If a reference is set, store the full path item in components
	// and return a path item $ref object.
	if store != nil && item.Reference != "" {
		if p, ok := store.GetPathItem(item.Reference); ok {
			// Skip registering operations for stored reference
			return p
		}
		for method, opt := range p.operations {
			AddOperation(store, item, method, opt)
		}
		return store.SetPathItem(item.Reference, item)
	}
	for method, opt := range p.operations {
		AddOperation(store, item, method, opt)
	}
	return item
}

// GetHandlers returns the registered handlers for each method on p.
// If no method is present the handler is registered for all supported methods.
func GetHandlers(p *Path) map[string]httpx.HandlerFunc {
	return p.handlers
}

// AddOperation applies opt to item for the given HTTP method.
// An empty method registers the operation for all supported methods.
func AddOperation(store *store.Store, item *v3.PathItem, method string, opt op.Option) {
	operation := &v3.Operation{}
	opt(operation, store)

	switch method {
	case http.MethodGet:
		item.Get = operation
	case http.MethodPost:
		item.Post = operation
	case http.MethodPut:
		item.Put = operation
	case http.MethodPatch:
		item.Patch = operation
	case http.MethodDelete:
		item.Delete = operation
	case "":
		item.Get = operation
		item.Post = operation
		item.Put = operation
		item.Patch = operation
		item.Delete = operation
	}
}
