// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"http"
	"path"
	"gorilla.googlecode.com/hg/gorilla/context"
)

// All error descriptions.
const (
	errRouteName        string = "Duplicated route name: %q."
	errVarName          string = "Duplicated route variable name: %q."
	// Template parsing.
	errUnbalancedBraces string = "Unbalanced curly braces in route template: %q."
	errBadTemplatePart  string = "Missing name or pattern in route template: %q."
	// URL building.
	errMissingRouteVar  string = "Missing route variable: %q."
	errBadRouteVar      string = "Route variable doesn't match: got %q, expected %q."
	errMissingHost      string = "Route doesn't have a host."
	errMissingPath      string = "Route doesn't have a path."
	// Empty parameter errors.
	errEmptyHost        string = "Host() requires a non-zero string, got %q."
	errEmptyPath        string = "Path() requires a non-zero string that starts with a slash, got %q."
	errEmptyPathPrefix  string = "PathPrefix() requires a non-zero string that starts with a slash, got %q."
	// Variadic errors.
	errEmptyHeaders     string = "Headers() requires at least a pair of parameters."
	errEmptyMethods     string = "Methods() requires at least one parameter."
	errEmptyQueries     string = "Queries() requires at least a pair of parameters."
	errEmptySchemes     string = "Schemes() requires at least one parameter."
	errOddHeaders       string = "Headers() requires an even number of parameters, got %v."
	errOddQueries       string = "Queries() requires an even number of parameters, got %v."
	errOddURLPairs      string = "URL() requires an even number of parameters, got %v."
)

// ----------------------------------------------------------------------------
// Context
// ----------------------------------------------------------------------------

// Vars stores the variables extracted from a URL.
type RouteVars map[string]string

// ctx is the request context namespace for this package.
//
// It stores route variables for each request.
var ctx = new(context.Namespace)

// Vars returns the route variables for the matched route in a given request.
func Vars(request *http.Request) RouteVars {
	rv := ctx.Get(request)
	if rv != nil {
		return rv.(RouteVars)
	}
	return nil
}

// ----------------------------------------------------------------------------
// Router
// ----------------------------------------------------------------------------

// Router registers routes to be matched and dispatches a handler.
//
// It implements the http.Handler interface, so it can be registered to serve
// requests. For example, using the default router:
//
//     func main() {
//         http.ListenAndServe(":8080", mux.DefaultRouter)
//     }
//
// Or, for Google App Engine, register it in a init() function:
//
//     func init() {
//         // Send all incoming requests to mux.DefaultRouter.
//         http.Handle("/", mux.DefaultRouter)
//     }
//
// The DefaultRouter is a Router instance ready to register URLs and handlers.
// If needed, new instances can be created and registered.
type Router struct {
	// Routes to be matched, in order.
	Routes      []*Route
	// Routes by name, for URL building.
	NamedRoutes map[string]*Route
	// Reference to the root router, where named routes are stored.
	rootRouter  *Router
}

// root returns the root router, where named routes are stored.
func (r *Router) root() *Router {
	if r.rootRouter == nil {
		return r
	}
	return r.rootRouter
}

// Match matches registered routes against the request.
func (r *Router) Match(request *http.Request) (rv *Route, ok bool) {
	for _, route := range r.Routes {
		if rv, ok = route.Match(request); ok {
			return
		}
	}
	return
}

// ServeHTTP dispatches the handler registered in the matched route.
//
// When there is a match, the route variables can be retrieved calling
// mux.Vars(request).
func (r *Router) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	// Clean path to canonical form and redirect.
	// (this comes from the http package)
	if p := cleanPath(request.URL.Path); p != request.URL.Path {
		writer.Header().Set("Location", p)
		writer.WriteHeader(http.StatusMovedPermanently)
		return
	}
	var handler http.Handler
	if route, ok := r.Match(request); ok {
		handler = route.GetHandler()
	}
	if handler == nil {
		handler = http.NotFoundHandler()
	}
	defer context.DefaultContext.Clear(request)
	handler.ServeHTTP(writer, request)
}

// Convenience route factories ------------------------------------------------

// NewRoute creates an empty route and registers it in the router.
func (r *Router) NewRoute() *Route {
	if r.Routes == nil {
		r.Routes = make([]*Route, 0)
	}
	route := newRoute()
	route.router = r
	r.Routes = append(r.Routes, route)
	return route
}

// Handler registers a new route and sets a handler.
//
// See also: Route.Handler().
func (r *Router) Handler(handler http.Handler) *Route {
	return r.NewRoute().Handler(handler)
}

// HandlerFunc registers a new route and sets a handler function.
//
// See also: Route.HandlerFunc().
func (r *Router) HandlerFunc(handler func(http.ResponseWriter,
							 *http.Request)) *Route {
	return r.NewRoute().HandlerFunc(handler)
}

// Handle registers a new route and sets a path and handler.
//
// See also: Route.Handle().
func (r *Router) Handle(path string, handler http.Handler) *Route {
	return r.NewRoute().Handle(path, handler)
}

// HandleFunc registers a new route and sets a path and handler function.
//
// See also: Route.HandleFunc().
func (r *Router) HandleFunc(path string, handler func(http.ResponseWriter,
							*http.Request)) *Route {
	return r.NewRoute().HandleFunc(path, handler)
}

// Name registers a new route and sets the route name.
//
// See also: Route.Name().
func (r *Router) Name(name string) *Route {
	return r.NewRoute().Name(name)
}

// Convenience route matcher factories ----------------------------------------

// Headers registers a new route and sets a headers matcher.
//
// See also: Route.Headers().
func (r *Router) Headers(pairs ...string) *Route {
	return r.NewRoute().Headers(pairs...)
}

// Host registers a new route and sets a host matcher.
//
// See also: Route.Host().
func (r *Router) Host(template string) *Route {
	return r.NewRoute().Host(template)
}

// Matcher registers a new route and sets a custom matcher function.
//
// See also: Route.Matcher().
func (r *Router) Matcher(matcherFunc MatcherFunc) *Route {
	return r.NewRoute().Matcher(matcherFunc)
}

// Methods registers a new route and sets a methods matcher.
//
// See also: Route.Methods().
func (r *Router) Methods(methods ...string) *Route {
	return r.NewRoute().Methods(methods...)
}

// Path registers a new route and sets a path matcher.
//
// See also: Route.Path().
func (r *Router) Path(template string) *Route {
	return r.NewRoute().Path(template)
}

// PathPrefix registers a new route and sets a path prefix matcher.
//
// See also: Route.PathPrefix().
func (r *Router) PathPrefix(template string) *Route {
	return r.NewRoute().PathPrefix(template)
}

// Queries registers a new route and sets a queries matcher.
//
// See also: Route.Queries().
func (r *Router) Queries(pairs ...string) *Route {
	return r.NewRoute().Queries(pairs...)
}

// Schemes registers a new route and sets a schemes matcher.
//
// See also: Route.Schemes().
func (r *Router) Schemes(schemes ...string) *Route {
	return r.NewRoute().Schemes(schemes...)
}

// ----------------------------------------------------------------------------
// Helpers
// ----------------------------------------------------------------------------

// cleanPath returns the canonical path for p, eliminating . and .. elements.
//
// Extracted from the http package.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}
