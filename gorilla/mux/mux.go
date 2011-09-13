// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"bytes"
	"fmt"
	"http"
	"path"
	"regexp"
	"strings"
	"url"
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

// NewRouter returns a new Router instance.
func NewRouter() *Router {
	return &Router{
		Routes:      make([]*Route, 0),
		NamedRoutes: make(map[string]*Route, 0),
	}
}

// root returns the root router, where named routes are stored.
func (r *Router) root() *Router {
	if r.rootRouter == nil {
		return r
	}
	return r.rootRouter
}

// Match matches registered routes against the request.
func (r *Router) Match(request *http.Request) (*Route, bool) {
	var rv *Route
	var ok bool
	for _, route := range r.Routes {
		if rv, ok = route.Match(request); ok {
			return rv, ok
		}
	}
	return nil, false
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
// Default Router
// ----------------------------------------------------------------------------

var DefaultRouter = NewRouter()
var NamedRoutes = DefaultRouter.NamedRoutes

// Convenience route factories ------------------------------------------------

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

// Convenience route matcher factories ----------------------------------------

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

// ----------------------------------------------------------------------------
// Route
// ----------------------------------------------------------------------------

// Route stores information to match a request.
type Route struct {
	// Reference to the router.
	router       *Router
	// Request handler for this route.
	handler      *http.Handler
	// List of matchers.
	matchers     []*routeMatcher
	// Special case matcher: parsed template for host matching.
	hostTemplate *parsedTemplate
	// Special case matcher: parsed template for path matching.
	pathTemplate *parsedTemplate
	// TODO (maybe)
	// Redirect access from paths not ending with slash to the slash'ed path
	// if the Route paths ends with a slash, and vice-versa.
	// http.Handle does it one way only: if pattern is /tree/, insert
	// permanent redirect for /tree.
	// strictSlash  bool
}

// newRoute returns a new Route instance.
func newRoute() *Route {
	return &Route{
		matchers: make([]*routeMatcher, 0),
	}
}

// GetHandler returns the handler registered in the route.
func (r *Route) GetHandler() http.Handler {
	return *r.handler
}

// Clone clones a route.
func (r *Route) Clone() *Route {
	// Fields are private and not changed once set, so we can reuse matchers
	// and parsed templates. Must make a copy of the matchers array, though.
	matchers := make([]*routeMatcher, len(r.matchers))
	copy(matchers, r.matchers)
	return &Route{
		router:       r.router,
		handler:      r.handler,
		matchers:     matchers,
		hostTemplate: r.hostTemplate,
		pathTemplate: r.pathTemplate,
	}
}

// Match matches this route against the request.
//
// It sets variables from the matched route in the context, if any.
func (r *Route) Match(req *http.Request) (*Route, bool) {
	// TODO Match the path unescaped?
	/*
	if path, ok := url.URLUnescape(r.URL.Path); ok {
		// URLUnescape converts '+' into ' ' (space). Revert this.
		path = strings.Replace(path, " ", "+", -1)
	} else {
		path = r.URL.Path
	}
	*/
	var hostMatches, pathMatches []string
	if r.hostTemplate != nil {
		hostMatches = r.hostTemplate.Regexp.FindStringSubmatch(req.URL.Host)
		if hostMatches == nil {
			return nil, false
		}
	}
	if r.pathTemplate != nil {
		pathMatches = r.pathTemplate.Regexp.FindStringSubmatch(req.URL.Path)
		if pathMatches == nil {
			return nil, false
		}
	}
	route := r
	if r.matchers != nil {
		var ok bool
		var rv *Route
		for _, matcher := range r.matchers {
			if rv, ok = (*matcher).Match(req); !ok {
				return nil, false
			} else if rv != nil {
				route = rv
				break
			}
		}
	}
	// We have a match.
	vars := make(RouteVars)
	if hostMatches != nil {
		for k, v := range r.hostTemplate.VarsN {
			vars[v] = hostMatches[k+1]
		}
	}
	if pathMatches != nil {
		for k, v := range r.pathTemplate.VarsN {
			vars[v] = pathMatches[k+1]
		}
	}
	ctx.Set(req, vars)
	return route, true
}

// Subrouting -----------------------------------------------------------------

// NewRouter creates a new router and adds it as a matcher for this route.
//
// This is used for subrouting: it will test the inner routes if other
// matchers matched. For example:
//
//     subrouter := mux.Host("www.domain.com").NewRouter()
//     subrouter.HandleFunc("/products/", ProductsHandler)
//     subrouter.HandleFunc("/products/{key}", ProductHandler)
//     subrouter.HandleFunc("/articles/{category}/{id:[0-9]+}"),
//                          ArticleHandler)
//
// In this example, the routes registered in the subrouter will only be tested
// if the host matches.
func (r *Route) NewRouter() *Router {
	router := &Router{
		Routes:     make([]*Route, 0),
		rootRouter: r.router.root(),
	}
	r.addMatcher(router)
	return router
}

// URL building ---------------------------------------------------------------

// URL builds a URL for the route.
//
// It accepts a sequence of key/value pairs for the route variables. For
// example, given this route:
//
//     mux.HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//         Name("article")
//
// ...a URL for it can be built using:
//
//     url := mux.NamedRoutes["article"].URL("category", "technology",
//                                           "id", "42")
//
// ...which will return an url.URL with the following path:
//
//     "/articles/technology/42"
//
// This also works for host variables:
//
//     mux.Host("{subdomain}.domain.com").
//              HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).
//              Name("article")
//
//     // url.String() will be "http://news.domain.com/articles/technology/42"
//     url := mux.NamedRoutes["article"].URL("subdomain", "news",
//                                           "category", "technology",
//                                           "id", "42")
//
// All variable names defined in the route are required, and their values must
// conform to the corresponding patterns, if any.
func (r *Route) URL(pairs ...string) *url.URL {
	var scheme, host, path string
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	if r.hostTemplate != nil {
		// Set a default scheme.
		scheme = "http"
		host = reverseRoute(r.hostTemplate, values)
	}
	if r.pathTemplate != nil {
		path = reverseRoute(r.pathTemplate, values)
	}
	return &url.URL{
		Scheme: scheme,
		Host:   host,
		Path:   path,
	}
}

// URLHost builds the host part of the URL for a route.
//
// The route must have a host defined.
func (r *Route) URLHost(pairs ...string) *url.URL {
	if r.hostTemplate == nil {
		panic(errMissingHost)
	}
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	return &url.URL{
		Scheme: "http",
		Host:   reverseRoute(r.hostTemplate, values),
	}
}

// URLPath builds the path part of the URL for a route.
//
// The route must have a path defined.
func (r *Route) URLPath(pairs ...string) *url.URL {
	if r.pathTemplate == nil {
		panic(errMissingPath)
	}
	values := stringMapFromPairs(errOddURLPairs, pairs...)
	return &url.URL{
		Path: reverseRoute(r.pathTemplate, values),
	}
}

// reverseRoute builds a URL part based on the route's parsed template.
func reverseRoute(tpl *parsedTemplate, values map[string]string) string {
	var value string
	var ok bool
	urlValues := make([]interface{}, len(tpl.VarsN))
	for k, v := range tpl.VarsN {
		if value, ok = values[v]; !ok {
			panic(fmt.Sprintf(errMissingRouteVar, v))
		}
		urlValues[k] = value
	}
	url := fmt.Sprintf(tpl.Reverse, urlValues...)
	if !tpl.Regexp.MatchString(url) {
		// The URL is checked against the full regexp, instead of checking
		// individual variables. This is faster but to provide a good error
		// message, we check individual regexps if the URL doesn't match.
		for k, v := range tpl.VarsN {
			if !tpl.VarsR[k].MatchString(values[v]) {
				panic(fmt.Sprintf(errBadRouteVar, values[v],
								  tpl.VarsR[k].String()))
			}
		}
	}
	return url
}

// Route predicates -----------------------------------------------------------

// Handler sets a handler for the route.
func (r *Route) Handler(handler http.Handler) *Route {
	r.handler = &handler
	return r
}

// HandlerFunc sets a handler function for the route.
func (r *Route) HandlerFunc(handler func(http.ResponseWriter,
							*http.Request)) *Route {
	return r.Handler(http.HandlerFunc(handler))
}

// Handle sets a path and handler for the route.
func (r *Route) Handle(path string, handler http.Handler) *Route {
	return r.Path(path).Handler(handler)
}

// HandleFunc sets a path and handler function for the route.
func (r *Route) HandleFunc(path string, handler func(http.ResponseWriter,
						   *http.Request)) *Route {
	return r.Path(path).Handler(http.HandlerFunc(handler))
}

// Name sets the route name, used to build URLs.
//
// A name must be unique for a router. It will panic if the same name was
// registered already.
func (r *Route) Name(name string) *Route {
	router := r.router.root()
	if _, ok := router.NamedRoutes[name]; ok {
		panic(fmt.Sprintf(errRouteName, name))
	}
	router.NamedRoutes[name] = r
	return r
}

// Route matchers -------------------------------------------------------------

// addMatcher adds a matcher to the array of route matchers.
func (r *Route) addMatcher(m routeMatcher) *Route {
	r.matchers = append(r.matchers, &m)
	return r
}

// Headers adds a matcher to match the request against header values.
//
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     mux.Headers("Content-Type", "application/json",
//                 "X-Requested-With", "XMLHttpRequest")
//
// The above route will only match if both request header values match.
func (r *Route) Headers(pairs ...string) *Route {
	headers := stringMapFromPairs(errOddHeaders, pairs...)
	if len(headers) == 0 {
		panic(errEmptyHeaders)
	}
	return r.addMatcher(&headerMatcher{headers: headers})
}

// Host adds a matcher to match the request against the URL host.
//
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next dot.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     mux.Host("www.domain.com")
//     mux.Host("{subdomain}.domain.com")
//     mux.Host("{subdomain:[a-z]+}.domain.com")
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Host(template string) *Route {
	if template == "" {
		panic(fmt.Sprintf(errEmptyHost, template))
	}
	r.hostTemplate = parseTemplate(template, "[^.]+", false,
								   variableNames(r.pathTemplate))
	return r
}

// Matcher adds a matcher to match the request using a custom function.
func (r *Route) Matcher(matcherFunc MatcherFunc) *Route {
	return r.addMatcher(&customMatcher{matcherFunc: matcherFunc})
}

// Methods adds a matcher to match the request against HTTP methods.
//
// It accepts a sequence of one or more methods to be matched, e.g.:
// "GET", "POST", "PUT".
func (r *Route) Methods(methods ...string) *Route {
	if len(methods) == 0 {
		panic(errEmptyMethods)
	}
	for k, v := range methods {
		methods[k] = strings.ToUpper(v)
	}
	return r.addMatcher(&methodMatcher{methods: methods})
}

// Path adds a matcher to match the request against the URL path.
//
// It accepts a template with zero or more URL variables enclosed by {}.
// Variables can define an optional regexp pattern to me matched:
//
// - {name} matches anything until the next slash.
//
// - {name:pattern} matches the given regexp pattern.
//
// For example:
//
//     mux.Path("/products/").Handler(ProductsHandler)
//     mux.Path("/products/{key}").Handler(ProductsHandler)
//     mux.Path("/articles/{category}/{id:[0-9]+}").
//             Handler(ArticleHandler)
//
// Variable names must be unique in a given route. They can be retrieved
// calling mux.Vars(request).
func (r *Route) Path(template string) *Route {
	if template == "" || template[0] != '/' {
		panic(fmt.Sprintf(errEmptyPath, template))
	}
	r.pathTemplate = parseTemplate(template, "[^/]+", false,
								   variableNames(r.hostTemplate))
	return r
}

// PathPrefix adds a matcher to match the request against a URL path prefix.
func (r *Route) PathPrefix(template string) *Route {
	if template == "" || template[0] != '/' {
		panic(fmt.Sprintf(errEmptyPathPrefix, template))
	}
	r.pathTemplate = parseTemplate(template, "[^/]+", true,
								   variableNames(r.pathTemplate))
	return r
}

// Queries adds a matcher to match the request against URL query values.
//
// It accepts a sequence of key/value pairs to be matched. For example:
//
//     mux.Queries("foo", "bar",
//                 "baz", "ding")
//
// The above route will only match if the URL contains the defined queries
// values, e.g.: ?foo=bar&baz=ding.
func (r *Route) Queries(pairs ...string) *Route {
	queries := stringMapFromPairs(errOddQueries, pairs...)
	if len(queries) == 0 {
		panic(errEmptyQueries)
	}
	return r.addMatcher(&queryMatcher{queries: queries})
}

// Schemes adds a matcher to match the request against URL schemes.
//
// It accepts a sequence of one or more schemes to be matched, e.g.:
// "http", "https".
func (r *Route) Schemes(schemes ...string) *Route {
	if len(schemes) == 0 {
		panic(errEmptySchemes)
	}
	for k, v := range schemes {
		schemes[k] = strings.ToLower(v)
	}
	return r.addMatcher(&schemeMatcher{schemes: schemes})
}

// ----------------------------------------------------------------------------
// Matchers
// ----------------------------------------------------------------------------

// MatcherFunc is the type used by custom matchers.
type MatcherFunc func(*http.Request) bool

// routeMatcher is the interface used by the router, route and route matchers.
//
// Only Router and Route actually return a route; it indicates a final match.
// Route matchers return nil and the result from the individual match.
type routeMatcher interface {
	Match(*http.Request) (*Route, bool)
}

// customMatcher matches the request using a custom matcher function.
type customMatcher struct {
	matcherFunc MatcherFunc
}

func (m *customMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, m.matcherFunc(request)
}

// headerMatcher matches the request against header values.
type headerMatcher struct {
	headers map[string]string
}

func (m *headerMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchMap(m.headers, request.Header, true)
}

// methodMatcher matches the request against HTTP methods.
type methodMatcher struct {
	methods []string
}

func (m *methodMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchInArray(m.methods, request.Method)
}

// queryMatcher matches the request against URL queries.
type queryMatcher struct {
	queries map[string]string
}

func (m *queryMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchMap(m.queries, request.URL.Query(), false)
}

// schemeMatcher matches the request against URL schemes.
type schemeMatcher struct {
	schemes []string
}

func (m *schemeMatcher) Match(request *http.Request) (*Route, bool) {
	return nil, matchInArray(m.schemes, request.URL.Scheme)
}

// ----------------------------------------------------------------------------
// Template parsing
// ----------------------------------------------------------------------------

// parsedTemplate stores a regexp and variables info for a route matcher.
type parsedTemplate struct {
	// Expanded regexp.
	Regexp  *regexp.Regexp
	// Reverse template.
	Reverse string
	// Variable names.
	VarsN   []string
	// Variable regexps (validators).
	VarsR   []*regexp.Regexp
}

// parseTemplate parses a route template, expanding variables into regexps.
//
// It will extract named variables, assemble a regexp to be matched, create
// a "reverse" template to build URLs and compile regexps to validate variable
// values used in URL building.
func parseTemplate(template string, defaultPattern string,
				   prefix bool, names *[]string) *parsedTemplate {
	// TODO Previously we accepted only Python-like identifiers for variable
	// names ([a-zA-Z_][a-zA-Z0-9_]*), but should we care at all?
	// Currently the only restriction is that name and pattern can't be empty.
	var raw, name, patt string
	var end int
	var parts []string
	pattern := bytes.NewBufferString("^")
	reverse := bytes.NewBufferString("")
	idxs := findAllVariableIndex(template)
	size := len(idxs)
	varsN := make([]string, size/2)
	varsR := make([]*regexp.Regexp, size/2)
	for i := 0; i < size; i += 2 {
		// 1. Set all values we are interested in.
		raw = template[end:idxs[i]]
		end = idxs[i+1]
		parts = strings.SplitN(template[idxs[i]+1:end-1], ":", 2)
		name = parts[0]
		if len(parts) == 1 {
			patt = defaultPattern
		} else {
			patt = parts[1]
		}
		// Name or pattern can't be empty.
		if name == "" || patt == "" {
			panic(fmt.Sprintf(errBadTemplatePart, template[idxs[i]:end]))
		}
		// Name must be unique for the route.
		if names != nil {
			if matchInArray(*names, name) {
				panic(fmt.Sprintf(errVarName, name))
			}
			*names = append(*names, name)
		}
		// 2. Build the regexp pattern.
		fmt.Fprintf(pattern, "%s(%s)", regexp.QuoteMeta(raw), patt)
		// 3. Build the reverse template.
		fmt.Fprintf(reverse, "%s%%s", raw)
		// 4. Append variable name and compiled pattern.
		varsN[i/2] = name
		varsR[i/2] = regexp.MustCompile(fmt.Sprintf("^%s$", patt))
	}
	// 5. Add the remaining.
	raw = template[end:]
	reverse.WriteString(raw)
	pattern.WriteString(regexp.QuoteMeta(raw))
	if !prefix {
		pattern.WriteString("$")
	}
	// Done!
	return &parsedTemplate{
		Regexp:  regexp.MustCompile(pattern.String()),
		Reverse: reverse.String(),
		VarsN:   varsN,
		VarsR:   varsR,
	}
}

// findAllVariableIndex returns index bounds for route template variables.
//
// It will panic if there are unbalanced curly braces.
func findAllVariableIndex(s string) []int {
	var level, idx int
	idxs := make([]int, 0)
	size := len(s)
	for i := 0; i < size; i++ {
		switch s[i] {
		case '{':
			if level++; level == 1 {
				idx = i
			}
		case '}':
			if level--; level == 0 {
				idxs = append(idxs, idx, i+1)
			} else if level < 0 {
				panic(fmt.Sprintf(errUnbalancedBraces, s))
			}
		}
	}
	if level != 0 {
		panic(fmt.Sprintf(errUnbalancedBraces, s))
	}
	return idxs
}

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
// Helpers
// ----------------------------------------------------------------------------

// stringMapFromPairs converts variadic string parameters to a string map.
func stringMapFromPairs(err string, pairs ...string) map[string]string {
	size := len(pairs)
	if size%2 != 0 {
		panic(fmt.Sprintf(err, pairs))
	}
	m := make(map[string]string, size/2)
	for i := 0; i < size; i += 2 {
		m[pairs[i]] = pairs[i+1]
	}
	return m
}

// variableNames returns a copy of variable names for route templates.
func variableNames(templates ...*parsedTemplate) *[]string {
	names := make([]string, 0)
	for _, t := range templates {
		if t != nil && len(t.VarsN) != 0 {
			names = append(names, t.VarsN...)
		}
	}
	return &names
}

// matchInArray returns true if the given string value is in the array.
func matchInArray(arr []string, value string) bool {
	for _, v := range arr {
		if v == value {
			return true
		}
	}
	return false
}

// matchMap returns true if the given key/value pairs exist in a given map.
func matchMap(toCheck map[string]string, toMatch map[string][]string,
			  canonicalKey bool) bool {
	var keyExists, valueExists bool
	var values []string
	var value string
	for k, v := range toCheck {
		// Check if key exists.
		if canonicalKey {
			k = http.CanonicalHeaderKey(k)
		}
		values, keyExists = toMatch[k]
		if !keyExists {
			return false
		}
		// Check if value exists.
		valueExists = false
		for _, value = range values {
			if v == value {
				valueExists = true
				break
			}
		}
		if !valueExists {
			return false
		}
	}
	return true
}

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
