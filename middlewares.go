package httpx

import (
	"net/http"
	"net/url"
	"strings"
)

// Unlike route middlewares (those that wrap the handler or are applied using Use)
// which are executed after ServeMux determines the route handler to execute,
// these middlewares are meant to wrap the ServeMux and intercept route matching if necessary.

// RewriteTrailingSlash rewrites all URLs ending with a trailing slash
// to the path equivalent without the trailing slash.
// This allows you to register routes without a trailing slash but not need to redirect
// if there was one in the URL.
//
// By default go http ServeMux redirects all non-trailing slash URLs to the URL with
// the trailing slash if two conditions are met:
//   - A route pattern for the non-trailing slash path is not registered.
//   - A route pattern for the trailing slash pattern is registered (either exact match or catch-all).
//
// We don't want this redirect to mess things up.
// Follow the rules below to prevent an infinite redirect loop caused by go http ServeMux redirects:
//   - For every route pattern ending with "/", register a route for the non-trailing slash path.
//     For example "/users/settings/" must be paired with "/users/settings"
//     to handle when exactly zero or one trailing slash is present in the URL.
//   - Treat route patterns ending with "/" as catch-all routes only.
//   - Avoid route patterns ending with "/{$}", prefer non-trailing slash pattern.
//
// The returned handler will not recognize route patterns that are meant to match URLs ending
// with a slash because such URLs will be matched with the non-trailing slash handler.
func RewriteTrailingSlash(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" || !strings.HasSuffix(r.URL.Path, "/") {
			handler.ServeHTTP(w, r)
			return
		}
		path := strings.TrimSuffix(r.URL.Path, "/")
		r2 := new(http.Request)
		*r2 = *r
		r2.URL = new(url.URL)
		*r2.URL = *r.URL
		r2.URL.Path, r2.URL.RawPath = path, path
		handler.ServeHTTP(w, r2)
	})
}
