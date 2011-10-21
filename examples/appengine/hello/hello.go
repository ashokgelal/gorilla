package hello

import (
	"fmt"
	"http"
	appengineSessions "gorilla.googlecode.com/hg/gorilla/appengine/sessions"
	"gorilla.googlecode.com/hg/gorilla/mux"
	"gorilla.googlecode.com/hg/gorilla/sessions"
)

func init() {
	// Register a couple of routes.
	mux.HandleFunc("/", homeHandler).Name("home")
	mux.HandleFunc("/{salutation}/{name}", helloHandler).Name("hello")
	mux.HandleFunc("/memcache-session", memcacheSessionHandler).Name("memcache-session")
	mux.HandleFunc("/datastore-session", datastoreSessionHandler).Name("datastore-session")

	// Send all incoming requests to mux.DefaultRouter.
	http.Handle("/", mux.DefaultRouter)

	// Register the datastore and memcache session stores.
	sessions.SetStore("datastore", new(appengineSessions.DatastoreSessionStore))
	sessions.SetStore("memcache", new(appengineSessions.MemcacheSessionStore))

	// Set secret keys for the session stores.
	sessions.SetStoreKeys("datastore",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))
	sessions.SetStoreKeys("memcache",
		[]byte("my-secret-key"),
		[]byte("1234567890123456"))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	url1 := mux.NamedRoutes["hello"].URL("salutation", "hello", "name", "world")
	url2 := mux.NamedRoutes["datastore-session"].URL()
	url3 := mux.NamedRoutes["memcache-session"].URL()
	fmt.Fprintf(w, "Try a <a href='%s'>hello</a>. Or a <a href='%s'>datastore</a> or <a href='%s'>memcache</a> session.", url1, url2, url3)
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	vars := mux.Vars(r)
	fmt.Fprintf(w, "%s, %s!", vars["salutation"], vars["name"])
}

func datastoreSessionHandler(w http.ResponseWriter, r *http.Request) {
	if session, err := sessions.Session(r, "mysession", "datastore"); err == nil {
		var msg string
		if value, ok := session["foo"]; ok {
			msg = fmt.Sprintf(`Value for session["foo"] is "%s".`, value)
		} else {
			session["foo"] = "bar"
			msg = fmt.Sprint(`No value set for session["foo"].`)
		}
		errors := sessions.Save(r, w)
		fmt.Fprintf(w, "%v\nErrors: %v.", msg, errors)
	} else {
		fmt.Fprintf(w, "Error getting session: %s", err)
	}
}

func memcacheSessionHandler(w http.ResponseWriter, r *http.Request) {
	if session, err := sessions.Session(r, "mysession", "memcache"); err == nil {
		var msg string
		if value, ok := session["foo"]; ok {
			msg = fmt.Sprintf(`Value for session["foo"] is "%s".`, value)
		} else {
			session["foo"] = "bar"
			msg = fmt.Sprint(`No value set for session["foo"].`)
		}
		errors := sessions.Save(r, w)
		fmt.Fprintf(w, "%v\nErrors: %v.", msg, errors)
	} else {
		fmt.Fprintf(w, "Error getting session: %s", err)
	}
}
