package openapi

import (
	"net/http"
	"strings"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/doc"
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/schema"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Router builds an OpenAPI document alongside route registration.
// Use NewRouter to create a standalone router, or UseRouter to extract one
// from a mux chain created with WithRouter.
type Router struct {
	mux      httpx.ServeMux
	doc      *v3.Document
	registry *schema.Registry
}

// NewRouter creates a standalone Router with the given document options.
// The router has no mux bound to it. Call WithMux before registering routes.
func NewRouter(opts ...doc.Option) *Router {
	schemas := orderedmap.New[string, *base.SchemaProxy]()
	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title:   "API",
			Version: "0.0.0",
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
		Components: &v3.Components{
			Schemas: schemas,
		},
	}
	for _, opt := range opts {
		opt(doc)
	}
	return &Router{nil, doc, schema.NewRegistry(schemas)}
}

// UseRouter extracts a Router from the mux chain created by WithRouter.
// The passed in mux is used to register routes, so any mux layered on
// top of WithRouter (Middleware, Prefix, Fallback, etc.) are still applied.
// Returns nil if no Router is found in the chain.
func UseRouter(mux httpx.ServeMux) *Router {
	originalMux := mux
	for {
		switch m := mux.(type) {
		case *openapiMux:
			return m.router.WithMux(originalMux)
		case httpx.Mux:
			mux = m.Mux()
		default:
			return nil
		}
	}
}

// WithMux returns a copy of the Router bound to mux.
// The document and schema registry are shared with the original Router.
func (r *Router) WithMux(mux httpx.ServeMux) *Router {
	return &Router{mux, r.doc, r.registry}
}

// Path records an OpenAPI operation for the given pattern.
// The pattern follows the same "METHOD /path" format as the Go
// ServeMux. If no method prefix is present the operation is recorded for all
// supported methods.
// Use this when the handler is registered separately.
func (r *Router) Path(pattern string, opt op.Option) {
	path, method := pattern, ""
	before, after, found := strings.Cut(path, " ")
	if found {
		method, path = before, after
	}
	path = normalizePath(httpx.MuxPrefix(r.mux) + path)

	item, ok := r.doc.Paths.PathItems.Get(path)
	if !ok {
		item = &v3.PathItem{}
		r.doc.Paths.PathItems.Set(path, item)
	}

	operation := &v3.Operation{}
	opt(operation, r.registry)

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
	default:
		if method == "" {
			item.Get = operation
			item.Post = operation
			item.Put = operation
			item.Patch = operation
			item.Delete = operation
		}
	}
}

// Route registers an OpenAPI operation and a route handler on the underlying mux.
// It combines Path and mux.Handle in one call.
func (r *Router) Route(pattern string, opt op.Option, handler func(http.ResponseWriter, *http.Request) error) {
	r.Path(pattern, opt)
	r.mux.Handle(pattern, httpx.Handler(r.mux, handler))
}

// GetDocument returns the underlying OpenAPI document.
func (r *Router) GetDocument() *v3.Document {
	return r.doc
}

// OpenAPIHandler returns an http.HandlerFunc that renders the OpenAPI document
// as JSON.
func (r *Router) OpenAPIHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		b, err := r.doc.RenderJSON("")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

// ReferenceHandler returns an http.HandlerFunc that serves a Scalar API reference UI.
// When options is nil the document is embedded directly in the HTML and the page title
// is set to the API title.
// Pass a non-nil *scalar.Options to override the default options.
func (r *Router) ReferenceHandler(options *scalar.Options) http.HandlerFunc {
	if options == nil {
		b, _ := r.doc.RenderJSON("")
		options = &scalar.Options{
			SpecContent: string(b),
			Theme:       scalar.ThemeDefault,
		}
	}
	if options.CustomOptions.PageTitle == "" {
		options.CustomOptions.PageTitle = r.doc.Info.Title
	}
	return func(w http.ResponseWriter, req *http.Request) {
		html, err := scalar.ApiReferenceHTML(options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(html))
	}
}

// normalizePath converts a Go ServeMux pattern path to a valid OpenAPI path.
// Named catch-all wildcards {name...} are trimmed to {name}, the exact-match
// marker {$} is dropped, and a trailing slash is replaced with {path} to
// represent the subtree pattern that matches everything beneath it.
func normalizePath(path string) string {
	segments := strings.Split(path, "/")
	out := segments[:0]
	for _, seg := range segments {
		if seg == "{$}" {
			continue
		}
		if strings.HasPrefix(seg, "{") && strings.HasSuffix(seg, "...}") {
			seg = seg[:len(seg)-4] + "}"
		}
		out = append(out, seg)
	}
	result := strings.Join(out, "/")
	if strings.HasSuffix(result, "/") {
		result += "{path}"
	}
	return result
}
