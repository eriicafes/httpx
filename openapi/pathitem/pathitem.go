package pathitem

import (
	"net/http"

	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/store"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

type Path struct {
	Opts       []Option
	Operations map[string]op.Option
	Handlers   map[string]httpx.HandlerFunc
}

func New(opts ...Option) *Path {
	return &Path{
		Opts:       opts,
		Operations: make(map[string]op.Option),
		Handlers:   make(map[string]httpx.HandlerFunc),
	}
}

func (p *Path) Operation(method string, opt op.Option) {
	p.Operations[method] = opt
}

func (p *Path) Route(method string, opt op.Option, handler func(http.ResponseWriter, *http.Request) error) {
	p.Operation(method, opt)
	p.Handlers[method] = handler
}

// Item constructs a v3.PathItem from the path's options and operations.
// If a reference is set, the full item is stored in components and a $ref object is returned.
func (p *Path) Item(store *store.Store) *v3.PathItem {
	item := &v3.PathItem{}
	for _, opt := range p.Opts {
		opt(item, store)
	}
	// If a reference is set, store the full path item in components
	// and return a path item $ref object.
	if store != nil && item.Reference != "" {
		if _, ok := store.GetPathItem(item.Reference); !ok {
			// Register operations only once for a *Path with reference
			for method, opt := range p.Operations {
				AddOperation(store, item, method, opt)
			}
		}
		return &v3.PathItem{Reference: store.SetPathItem(item.Reference, item)}
	}
	// Register operations always if *Path has no reference
	for method, opt := range p.Operations {
		AddOperation(store, item, method, opt)
	}
	return item
}

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
	case http.MethodOptions:
		item.Options = operation
	case http.MethodHead:
		item.Head = operation
	case http.MethodTrace:
		item.Trace = operation
	case "":
		item.Get = operation
		item.Post = operation
		item.Put = operation
		item.Patch = operation
		item.Delete = operation
		item.Options = operation
		item.Head = operation
		item.Trace = operation
	}
}
