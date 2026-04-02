// Package openapi builds an OpenAPI document alongside httpx route registration.
package openapi

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/MarceloPetrucio/go-scalar-api-reference"
	"github.com/eriicafes/httpx"
	"github.com/eriicafes/httpx/openapi/doc"
	"github.com/eriicafes/httpx/openapi/op"
	"github.com/eriicafes/httpx/openapi/pathitem"
	"github.com/eriicafes/httpx/openapi/store"
	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
	"github.com/pb33f/libopenapi/orderedmap"
)

// Router builds an OpenAPI document alongside route registration.
// Use NewRouter to create a standalone router, or UseRouter to extract one
// from a mux chain created with WithRouter.
type Router struct {
	mux   httpx.ServeMux
	doc   *v3.Document
	store *store.Store
}

// NewRouter creates a standalone Router with the given document options.
// The router has no mux bound to it. Call WithMux before registering routes.
func NewRouter(title, version string, opts ...doc.Option) *Router {
	components := &v3.Components{}
	store := store.New(components)
	doc := &v3.Document{
		Version: "3.1.0",
		Info: &base.Info{
			Title:   title,
			Version: version,
		},
		Paths: &v3.Paths{
			PathItems: orderedmap.New[string, *v3.PathItem](),
		},
		Components: components,
	}
	for _, opt := range opts {
		opt(doc, store)
	}
	return &Router{nil, doc, store}
}

// UseRouter extracts a Router from the mux chain created by WithRouter.
// The passed in mux is used to register routes, so any mux layered on
// top of WithRouter (Middleware, Prefix, Fallback, etc.) are still applied.
// Panics if no Router is found in the chain.
func UseRouter(mux httpx.ServeMux) *Router {
	originalMux := mux
	for {
		switch m := mux.(type) {
		case *openapiMux:
			return m.router.WithMux(originalMux)
		case httpx.Mux:
			mux = m.Mux()
		default:
			panic("httpx/openapi: no Router found in mux chain; wrap mux with openapi.WithRouter")
		}
	}
}

// WithMux returns a copy of the Router bound to mux.
// The document and schema registry are shared with the original Router.
func (r *Router) WithMux(mux httpx.ServeMux) *Router {
	return &Router{mux, r.doc, r.store}
}

// Operation registers an OpenAPI operation for the given pattern.
// The pattern follows the same "METHOD /path" format as the Go ServeMux.
// If no method is present the operation is registered for all supported methods.
// Use this when the handler is registered separately.
func (r *Router) Operation(pattern string, opt op.Option) {
	var method string
	before, after, found := strings.Cut(pattern, " ")
	if found {
		method, pattern = before, after
	}
	pattern = httpx.MuxPrefix(r.mux) + pattern
	path := normalizePath(pattern)

	item, ok := r.doc.Paths.PathItems.Get(path)
	if !ok {
		item = &v3.PathItem{}
		r.doc.Paths.PathItems.Set(path, item)
	}

	// Check for conflict (when existing item has a reference)
	if item.Reference != "" {
		panic(fmt.Errorf("httpx/openapi: path %q conflicts with existing path with reference %q", path, item.Reference))
	}
	pathitem.AddOperation(r.store, item, method, opt)
}

// Route registers an OpenAPI operation and a route handler on the underlying mux.
func (r *Router) Route(pattern string, opt op.Option, handler func(http.ResponseWriter, *http.Request) error) {
	r.Operation(pattern, opt)
	r.mux.Handle(pattern, httpx.Handler(r.mux, handler))
}

// Handle registers an OpenAPI operation and a handler on the underlying mux.
func (r *Router) Handle(pattern string, opt op.Option, handler http.Handler) {
	r.Operation(pattern, opt)
	r.mux.Handle(pattern, handler)
}

// HandleFunc registers an OpenAPI operation and a handler func on the underlying mux.
func (r *Router) HandleFunc(pattern string, opt op.Option, handler func(http.ResponseWriter, *http.Request)) {
	r.Operation(pattern, opt)
	r.mux.HandleFunc(pattern, handler)
}

// Path registers an OpenAPI pathitem and its registered handlers for each method.
func (r *Router) Path(pattern string, p *pathitem.Path) {
	pattern = httpx.MuxPrefix(r.mux) + pattern
	path := normalizePath(pattern)

	// Check for conflict (when existing or new item has a reference)
	existingItem, ok := r.doc.Paths.PathItems.Get(path)
	item := pathitem.GetPathItem(p, r.store, existingItem)
	if ok && existingItem.Reference != "" {
		panic(fmt.Errorf("httpx/openapi: path %q conflicts with existing path with reference %q", path, existingItem.Reference))
	}
	if ok && item.Reference != "" {
		panic(fmt.Errorf("httpx/openapi: path %q with reference %q conflicts with existing path", path, item.Reference))
	}
	r.doc.Paths.PathItems.Set(path, item)

	// Register handlers
	for method, handler := range pathitem.GetHandlers(p) {
		pattern := pattern
		if method != "" {
			pattern = method + " " + pattern
		}
		r.mux.Handle(pattern, httpx.Handler(r.mux, handler))
	}
}

// GetDocument returns the underlying OpenAPI document.
func (r *Router) GetDocument() *v3.Document {
	return r.doc
}

// OpenAPIJSONHandler returns an http.HandlerFunc that renders the OpenAPI document as JSON.
func (r *Router) OpenAPIJSONHandler() http.HandlerFunc {
	b, err := r.doc.RenderJSON("")
	return func(w http.ResponseWriter, req *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(b)
	}
}

// OpenAPIYAMLHandler returns an http.HandlerFunc that renders the OpenAPI document as YAML.
func (r *Router) OpenAPIYAMLHandler() http.HandlerFunc {
	b, err := r.doc.Render()
	return func(w http.ResponseWriter, req *http.Request) {
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/yaml")
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
	if result == "" {
		result = "/"
	}
	return result
}
