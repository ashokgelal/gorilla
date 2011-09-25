// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"http"
)

// ----------------------------------------------------------------------------
// Default Router
// ----------------------------------------------------------------------------
// There's no new logic here, only functions that mirror Router methods to
// change the default router instance.

// DefaultRouter is a default Router instance for convenience.
var DefaultRouter = new(Router)

// NamedRoutes is the DefaultRouter's NamedRoutes field.
var NamedRoutes = DefaultRouter.NamedRoutes

// AddRoute registers a route in the default router.
func AddRoute(route *Route) *Router {
	return DefaultRouter.AddRoute(route)
}

// Route factories ------------------------------------------------------------

// NewRoute creates an empty route and registers it in the router.
func NewRoute() *Route {
	return DefaultRouter.NewRoute()
}

// Handler registers a new route and sets a handler.
//
// See also: Route.Handler().
func Handler(handler http.Handler) *Route {
	return DefaultRouter.NewRoute().Handler(handler)
}

// HandlerFunc registers a new route and sets a handler function.
//
// See also: Route.HandlerFunc().
func HandlerFunc(handler func(http.ResponseWriter, *http.Request)) *Route {
	return DefaultRouter.NewRoute().HandlerFunc(handler)
}

// Handle registers a new route and sets a path and handler.
//
// See also: Route.Handle().
func Handle(path string, handler http.Handler) *Route {
	return DefaultRouter.NewRoute().Handle(path, handler)
}

// HandleFunc registers a new route and sets a path and handler function.
//
// See also: Route.HandleFunc().
func HandleFunc(path string, handler func(http.ResponseWriter,
				*http.Request)) *Route {
	return DefaultRouter.NewRoute().HandleFunc(path, handler)
}

// Name registers a new route and sets the route name.
//
// See also: Route.Name().
func Name(name string) *Route {
	return DefaultRouter.NewRoute().Name(name)
}

// Route matcher factories ----------------------------------------------------

// Headers registers a new route and sets a headers matcher.
//
// See also: Route.Headers().
func Headers(pairs ...string) *Route {
	return DefaultRouter.NewRoute().Headers(pairs...)
}

// Host registers a new route and sets a host matcher.
//
// See also: Route.Host().
func Host(template string) *Route {
	return DefaultRouter.NewRoute().Host(template)
}

// Matcher registers a new route and sets a custom matcher function.
//
// See also: Route.Matcher().
func Matcher(matcherFunc MatcherFunc) *Route {
	return DefaultRouter.NewRoute().Matcher(matcherFunc)
}

// Methods registers a new route and sets a methods matcher.
//
// See also: Route.Methods().
func Methods(methods ...string) *Route {
	return DefaultRouter.NewRoute().Methods(methods...)
}

// Path registers a new route and sets a path matcher.
//
// See also: Route.Path().
func Path(template string) *Route {
	return DefaultRouter.NewRoute().Path(template)
}

// PathPrefix registers a new route and sets a path prefix matcher.
//
// See also: Route.PathPrefix().
func PathPrefix(template string) *Route {
	return DefaultRouter.NewRoute().PathPrefix(template)
}

// Queries registers a new route and sets a queries matcher.
//
// See also: Route.Queries().
func Queries(pairs ...string) *Route {
	return DefaultRouter.NewRoute().Queries(pairs...)
}

// Schemes registers a new route and sets a schemes matcher.
//
// See also: Route.Schemes().
func Schemes(schemes ...string) *Route {
	return DefaultRouter.NewRoute().Schemes(schemes...)
}
