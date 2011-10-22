// Copyright 2011 Rodrigo Moraes. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mux

import (
	"http"
	"testing"
)

func TestRouteMatchers(t *testing.T) {
	var scheme, host, path, query, method string
	var headers map[string]string
	var resultVars map[bool]map[string]string

	router := new(Router)
	router.Host("{var1}.google.com").
		Path("/{var2:[a-z]+}/{var3:[0-9]+}").
		Queries("foo", "bar").
		Methods("GET").
		Schemes("https").
		Headers("x-requested-with", "XMLHttpRequest")
	router.Host("www.{var4}.com").
		PathPrefix("/foo/{var5:[a-z]+}/{var6:[0-9]+}").
		Queries("baz", "ding").
		Methods("POST").
		Schemes("http").
		Headers("Content-Type", "application/json")

	reset := func() {
		// Everything match.
		scheme = "https"
		host = "www.google.com"
		path = "/product/42"
		query = "?foo=bar"
		method = "GET"
		headers = map[string]string{"X-Requested-With": "XMLHttpRequest"}
		resultVars = map[bool]map[string]string{
			true:  map[string]string{"var1": "www", "var2": "product", "var3": "42"},
			false: map[string]string{},
		}
	}

	reset2 := func() {
		// Everything match.
		scheme = "http"
		host = "www.google.com"
		path = "/foo/product/42/path/that/is/ignored"
		query = "?baz=ding"
		method = "POST"
		headers = map[string]string{"Content-Type": "application/json"}
		resultVars = map[bool]map[string]string{
			true:  map[string]string{"var4": "google", "var5": "product", "var6": "42"},
			false: map[string]string{},
		}
	}

	match := func(shouldMatch bool) {
		url := scheme + "://" + host + path + query
		request, _ := http.NewRequest(method, url, nil)
		for key, value := range headers {
			request.Header.Add(key, value)
		}

		_, matched := router.Match(request)
		if matched != shouldMatch {
			// Need better messages. :)
			if matched {
				t.Errorf("Should match.")
			} else {
				t.Errorf("Should not match.")
			}
		}

		if matched {
			vars := Vars(request)
			expectedVars := resultVars[shouldMatch]
			if len(vars) != len(expectedVars) {
				t.Errorf("Expected vars: %v Got: %v.", expectedVars, vars)
			}
			for name, value := range vars {
				if expectedVars[name] != value {
					t.Errorf("Expected vars: %v Got: %v.", expectedVars, vars)
				}
			}
		}
	}

	// 1st route --------------------------------------------------------------

	// Everything match.
	reset()
	match(true)

	// Scheme doesn't match.
	reset()
	scheme = "http"
	match(false)

	// Host doesn't match.
	reset()
	host = "www.mygoogle.com"
	match(false)

	// Path doesn't match.
	reset()
	path = "/product/notdigits"
	match(false)

	// Query doesn't match.
	reset()
	query = "?foo=baz"
	match(false)

	// Method doesn't match.
	reset()
	method = "POST"
	match(false)

	// Header doesn't match.
	reset()
	headers = map[string]string{}
	match(false)

	// Everything match, again.
	reset()
	match(true)

	// 2nd route --------------------------------------------------------------

	// Everything match.
	reset2()
	match(true)

	// Scheme doesn't match.
	reset2()
	scheme = "https"
	match(false)

	// Host doesn't match.
	reset2()
	host = "sub.google.com"
	match(false)

	// Path doesn't match.
	reset2()
	path = "/bar/product/42"
	match(false)

	// Query doesn't match.
	reset2()
	query = "?foo=baz"
	match(false)

	// Method doesn't match.
	reset2()
	method = "GET"
	match(false)

	// Header doesn't match.
	reset2()
	headers = map[string]string{}
	match(false)

	// Everything match, again.
	reset2()
	match(true)
}

type headerMatcherTest struct {
	matcher *headerMatcher
	headers map[string]string
	result  bool
}

var headerMatcherTests = []headerMatcherTest{
	{
		matcher: &headerMatcher{map[string]string{"x-requested-with": "XMLHttpRequest"}},
		headers: map[string]string{"X-Requested-With": "XMLHttpRequest"},
		result:  true,
	},
	{
		matcher: &headerMatcher{map[string]string{"x-requested-with": ""}},
		headers: map[string]string{"X-Requested-With": "anything"},
		result:  true,
	},
	{
		matcher: &headerMatcher{map[string]string{"x-requested-with": "XMLHttpRequest"}},
		headers: map[string]string{},
		result:  false,
	},
}

type hostMatcherTest struct {
	matcher *Route
	url     string
	vars    RouteVars
	result  bool
}

var hostMatcherTests = []hostMatcherTest{
	{
		matcher: new(Router).Host("{foo:[a-z][a-z][a-z]}.{bar:[a-z][a-z][a-z]}.{baz:[a-z][a-z][a-z]}"),
		url:     "http://abc.def.ghi/",
		vars:    map[string]string{"foo": "abc", "bar": "def", "baz": "ghi"},
		result:  true,
	},
	{
		matcher: new(Router).Host("{foo:[a-z][a-z][a-z]}.{bar:[a-z][a-z][a-z]}.{baz:[a-z][a-z][a-z]}"),
		url:     "http://a.b.c/",
		vars:    map[string]string{"foo": "abc", "bar": "def", "baz": "ghi"},
		result:  false,
	},
}

type methodMatcherTest struct {
	matcher *methodMatcher
	method  string
	result  bool
}

var methodMatcherTests = []methodMatcherTest{
	{
		matcher: &methodMatcher{[]string{"GET", "POST", "PUT"}},
		method:  "GET",
		result:  true,
	},
	{
		matcher: &methodMatcher{[]string{"GET", "POST", "PUT"}},
		method:  "POST",
		result:  true,
	},
	{
		matcher: &methodMatcher{[]string{"GET", "POST", "PUT"}},
		method:  "PUT",
		result:  true,
	},
	{
		matcher: &methodMatcher{[]string{"GET", "POST", "PUT"}},
		method:  "DELETE",
		result:  false,
	},
}

type pathMatcherTest struct {
	matcher *Route
	url     string
	vars    RouteVars
	result  bool
}

var pathMatcherTests = []pathMatcherTest{
	{
		matcher: new(Router).Path("/{foo:[0-9][0-9][0-9]}/{bar:[0-9][0-9][0-9]}/{baz:[0-9][0-9][0-9]}"),
		url:     "http://localhost:8080/123/456/789",
		vars:    map[string]string{"foo": "123", "bar": "456", "baz": "789"},
		result:  true,
	},
	{
		matcher: new(Router).Path("/{foo:[0-9][0-9][0-9]}/{bar:[0-9][0-9][0-9]}/{baz:[0-9][0-9][0-9]}"),
		url:     "http://localhost:8080/1/2/3",
		vars:    map[string]string{"foo": "123", "bar": "456", "baz": "789"},
		result:  false,
	},
}

type queryMatcherTest struct {
	matcher *queryMatcher
	url     string
	result  bool
}

var queryMatcherTests = []queryMatcherTest{
	{
		matcher: &queryMatcher{map[string]string{"foo": "bar", "baz": "ding"}},
		url:     "http://localhost:8080/?foo=bar&baz=ding",
		result:  true,
	},
	{
		matcher: &queryMatcher{map[string]string{"foo": "", "baz": ""}},
		url:     "http://localhost:8080/?foo=anything&baz=anything",
		result:  true,
	},
	{
		matcher: &queryMatcher{map[string]string{"foo": "ding", "baz": "bar"}},
		url:     "http://localhost:8080/?foo=bar&baz=ding",
		result:  false,
	},
	{
		matcher: &queryMatcher{map[string]string{"bar": "foo", "ding": "baz"}},
		url:     "http://localhost:8080/?foo=bar&baz=ding",
		result:  false,
	},
}

type schemeMatcherTest struct {
	matcher *schemeMatcher
	url     string
	result  bool
}

var schemeMatcherTests = []schemeMatcherTest{
	{
		matcher: &schemeMatcher{[]string{"http", "https"}},
		url:     "http://localhost:8080/",
		result:  true,
	},
	{
		matcher: &schemeMatcher{[]string{"http", "https"}},
		url:     "https://localhost:8080/",
		result:  true,
	},
	{
		matcher: &schemeMatcher{[]string{"https"}},
		url:     "http://localhost:8080/",
		result:  false,
	},
	{
		matcher: &schemeMatcher{[]string{"http"}},
		url:     "https://localhost:8080/",
		result:  false,
	},
}

type urlBuildingTest struct {
	route *Route
	vars  []string
	url   string
}

var urlBuildingTests = []urlBuildingTest{
	{
		route: newRoute().Host("foo.domain.com"),
		vars:  []string{},
		url:   "http://foo.domain.com",
	},
	{
		route: newRoute().Host("{subdomain}.domain.com"),
		vars:  []string{"subdomain", "bar"},
		url:   "http://bar.domain.com",
	},
	{
		route: newRoute().Host("foo.domain.com").Path("/articles"),
		vars:  []string{},
		url:   "http://foo.domain.com/articles",
	},
	{
		route: newRoute().Path("/articles"),
		vars:  []string{},
		url:   "/articles",
	},
	{
		route: newRoute().Path("/articles/{category}/{id:[0-9]+}"),
		vars:  []string{"category", "technology", "id", "42"},
		url:   "/articles/technology/42",
	},
	{
		route: newRoute().Host("{subdomain}.domain.com").Path("/articles/{category}/{id:[0-9]+}"),
		vars:  []string{"subdomain", "foo", "category", "technology", "id", "42"},
		url:   "http://foo.domain.com/articles/technology/42",
	},
}

func TestHeaderMatcher(t *testing.T) {
	for _, v := range headerMatcherTests {
		request, _ := http.NewRequest("GET", "http://localhost:8080/", nil)
		for key, value := range v.headers {
			request.Header.Add(key, value)
		}
		_, result := v.matcher.Match(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, request.Header)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, request.Header)
			}
		}
	}
}

func TestHostMatcher(t *testing.T) {
	for _, v := range hostMatcherTests {
		request, _ := http.NewRequest("GET", v.url, nil)
		_, result := v.matcher.Match(request)
		vars := Vars(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, v.url)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, v.url)
			}
		}
		if result {
			if len(vars) != len(v.vars) {
				t.Errorf("%#v: vars length should be %v, got %v.", v.matcher, len(v.vars), len(vars))
			}
			for name, value := range vars {
				if v.vars[name] != value {
					t.Errorf("%#v: expected value %v for key %v, got %v.", v.matcher, v.vars[name], name, value)
				}
			}
		} else {
			if len(vars) != 0 {
				t.Errorf("%#v: vars length should be 0, got %v.", v.matcher, len(vars))
			}
		}
	}
}

func TestMethodMatcher(t *testing.T) {
	for _, v := range methodMatcherTests {
		request, _ := http.NewRequest(v.method, "http://localhost:8080/", nil)
		_, result := v.matcher.Match(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, v.method)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, v.method)
			}
		}
	}
}

func TestPathMatcher(t *testing.T) {
	for _, v := range pathMatcherTests {
		request, _ := http.NewRequest("GET", v.url, nil)
		_, result := v.matcher.Match(request)
		vars := Vars(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, v.url)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, v.url)
			}
		}
		if result {
			if len(vars) != len(v.vars) {
				t.Errorf("%#v: vars length should be %v, got %v.", v.matcher, len(v.vars), len(vars))
			}
			for name, value := range vars {
				if v.vars[name] != value {
					t.Errorf("%#v: expected value %v for key %v, got %v.", v.matcher, v.vars[name], name, value)
				}
			}
		} else {
			if len(vars) != 0 {
				t.Errorf("%#v: vars length should be 0, got %v.", v.matcher, len(vars))
			}
		}
	}
}

func TestQueryMatcher(t *testing.T) {
	for _, v := range queryMatcherTests {
		request, _ := http.NewRequest("GET", v.url, nil)
		_, result := v.matcher.Match(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, v.url)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, v.url)
			}
		}
	}
}

func TestSchemeMatcher(t *testing.T) {
	for _, v := range queryMatcherTests {
		request, _ := http.NewRequest("GET", v.url, nil)
		_, result := v.matcher.Match(request)
		if result != v.result {
			if v.result {
				t.Errorf("%#v: should match %v.", v.matcher, v.url)
			} else {
				t.Errorf("%#v: should not match %v.", v.matcher, v.url)
			}
		}
	}
}

func TestUrlBuilding(t *testing.T) {
	for _, v := range urlBuildingTests {
		url := v.route.URL(v.vars...).String()
		if url != v.url {
			reversePath := ""
			reverseHost := ""
			if v.route.pathTemplate != nil {
				reversePath = v.route.pathTemplate.Reverse
			}
			if v.route.hostTemplate != nil {
				reverseHost = v.route.hostTemplate.Reverse
			}

			t.Errorf("%#v:\nexpected: %q\ngot: %q\nreverse path: %q\nreverse host: %q", v.route, v.url, url, reversePath, reverseHost)
		}
	}

	ArticleHandler := func(w http.ResponseWriter, r *http.Request) {
	}

	HandleFunc("/articles/{category}/{id:[0-9]+}", ArticleHandler).Name("article")

	url := NamedRoutes["article"].URL("category", "technology", "id", "42")
	expected := "/articles/technology/42"
	if url.String() != expected {
		t.Errorf("Expected %v, got %v", expected, url.String())
	}
}

func TestSubRouting(t *testing.T) {
	// Example from docs.
	router := new(Router)
	subrouter := router.Host("www.domain.com").NewRouter()
	route := subrouter.Path("/products/").Name("products")

	url := "http://www.domain.com/products/"
	request, _ := http.NewRequest("GET", url, nil)
	rv, ok := router.Match(request)

	if !ok || rv.Route != route {
		t.Errorf("Expectd same route, got %+v.", rv.Route)
	}

	builtUrl := router.NamedRoutes["products"].URL().String()
	if builtUrl != "/products/" {
		t.Errorf("Expectd %q, got %q.", url, builtUrl)
	}
}

func TestVariableNames(t *testing.T) {
	route := newRoute().Host("{arg1}.domain.com").Path("/{arg2}/{arg3:[0-9]+}")
	names := variableNames(route.hostTemplate, route.pathTemplate)
	if len(*names) != 3 {
		t.Errorf("Expected %v variable names in %v.", len(*names), *names)
	}
}

/*
func TestNewRegexp(t *testing.T) {
	var p *parsedTemplate
	var matches []string

	tests := map[string]map[string][]string{
		"/{foo:a{2}}": {
			"/a":    nil,
			"/aa":   {"aa"},
			"/aaa":  nil,
			"/aaaa": nil,
		},
		"/{foo:a{2,}}": {
			"/a":    nil,
			"/aa":   {"aa"},
			"/aaa":  {"aaa"},
			"/aaaa": {"aaaa"},
		},
		"/{foo:a{2,3}}": {
			"/a":    nil,
			"/aa":   {"aa"},
			"/aaa":  {"aaa"},
			"/aaaa": nil,
		},
		"/{foo:[a-z]{3}}/{bar:[a-z]{2}}": {
			"/a":       nil,
			"/ab":      nil,
			"/abc":     nil,
			"/abcd":    nil,
			"/abc/ab":  {"abc", "ab"},
			"/abc/abc": nil,
			"/abcd/ab": nil,
		},
		`/{foo:\w{3,}}/{bar:\d{2,}}`: {
			"/a":        nil,
			"/ab":       nil,
			"/abc":      nil,
			"/abc/1":    nil,
			"/abc/12":   {"abc", "12"},
			"/abcd/12":  {"abcd", "12"},
			"/abcd/123": {"abcd", "123"},
		},
	}

	for pattern, paths := range tests {
		p, _ = parseTemplate(pattern, "[^/]+", false, nil)
		for path, result := range paths {
			matches = p.Regexp.FindStringSubmatch(path)
			if result == nil {
				if matches != nil {
					t.Errorf("%v should not match %v.", pattern, path)
				}
			} else {
				if len(matches) != len(result) + 1 {
					t.Errorf("Expected %v matches, got %v.", len(result)+1 , len(matches))
				} else {
					for k, v := range result {
						if matches[k+1] != v {
							t.Errorf("Expected %v, got %v.", v, matches[k+1])
						}
					}
				}
			}
		}
	}
}
*/
