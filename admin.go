package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
)

func newAPIHandler(path string, safeNetworks []*net.IPNet) http.Handler {
	return &admin{
		store: NewStore(path),
		safe:  safeNetworks,
	}
}

type admin struct {
	store *Store
	safe  []*net.IPNet
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := a.allowed(r); err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var t rest
	switch id := getID(r); id {
	case "/", "":
		t = &collection{store: a.store}
	default:
		t = &item{name: id, store: a.store}
	}

	switch r.Method {
	case "GET":
		t.get(w, r)
	case "PUT":
		t.put(w, r)
	case "POST":
		t.post(w, r)
	case "DELETE":
		t.del(w, r)
	case "PATCH":
		t.patch(w, r)
	case "OPTIONS":
		t.options(w, r)
	case "HEAD":
		t.head(w, r)
	default:
		t.def(w, r)
	}
}

// allowed decides if a request is permitted
func (a *admin) allowed(r *http.Request) error {
	lip := lastForwarder(r)
	if a != nil {
		// check if the client ip is in the "safe" networks
		ip := net.ParseIP(lip)
		for _, s := range a.safe {
			if s.Contains(ip) {
				return nil
			}
		}
		// possibly fallback to some sort of authentication scheme
		// TODO
	}
	return fmt.Errorf("request not permitted. [%s]", lip)
}

type rest interface {
	get(http.ResponseWriter, *http.Request)
	put(http.ResponseWriter, *http.Request)
	post(http.ResponseWriter, *http.Request)
	del(http.ResponseWriter, *http.Request)
	patch(http.ResponseWriter, *http.Request)
	options(http.ResponseWriter, *http.Request)
	head(http.ResponseWriter, *http.Request)
	def(http.ResponseWriter, *http.Request)
}

type collection struct {
	store *Store
}

func (c *collection) get(w http.ResponseWriter, r *http.Request) {
	prefix := r.URL.Query().Get("prefix")
	switch r, err := c.store.List(prefix); err {
	case nil:
		writeJSON(w, r)
	case errNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *collection) post(w http.ResponseWriter, r *http.Request) {
	var repo Repo
	if err := readJSON(r, &repo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	switch err := c.store.Create(repo); err {
	case nil:
		w.Header().Set("Location", repo.ImportRoot)
		w.WriteHeader(http.StatusCreated)
	case errConflict:
		http.Error(w, "already exists", http.StatusConflict)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (c *collection) put(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "use resource id", http.StatusMethodNotAllowed)
}

func (c *collection) del(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "use resource id", http.StatusMethodNotAllowed)
}

func (c *collection) patch(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "use resource id", http.StatusMethodNotAllowed)
}

func (c *collection) options(w http.ResponseWriter, r *http.Request) {}
func (c *collection) head(w http.ResponseWriter, r *http.Request)    {}
func (c *collection) def(w http.ResponseWriter, r *http.Request)     {}

type item struct {
	name  string
	store *Store
}

func (i *item) get(w http.ResponseWriter, r *http.Request) {
	switch repos, err := i.store.Read(i.name); err {
	case nil:
		writeJSON(w, repos)
	case errNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func (i *item) put(w http.ResponseWriter, r *http.Request)  {}
func (i *item) post(w http.ResponseWriter, r *http.Request) {}

func (i *item) del(w http.ResponseWriter, r *http.Request) {
	switch err := i.store.Delete(i.name); err {
	case nil:
	case errNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (i *item) patch(w http.ResponseWriter, r *http.Request)   {}
func (i *item) options(w http.ResponseWriter, r *http.Request) {}
func (i *item) head(w http.ResponseWriter, r *http.Request)    {}
func (i *item) def(w http.ResponseWriter, r *http.Request)     {}

func getID(r *http.Request) string {
	if r == nil {
		return ""
	}

	return r.URL.Path
}

func readJSON(r *http.Request, v interface{}) error {
	if r == nil || r.Body == nil {
		return errors.New("invalid request")
	}
	if ct := r.Header.Get("Content-Type"); !strings.HasPrefix(ct, "application/json") {
		return fmt.Errorf("invalid content-type: %q", ct)
	}
	defer r.Body.Close()
	return json.NewDecoder(r.Body).Decode(v)
}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error writing response: %v", err)
	}
}
