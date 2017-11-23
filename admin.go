package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
)

func newAPIHandler(path string, safeIPs []string) http.Handler {
	var safe []*net.IPNet
	for _, s := range safeIPs {
		_, n, err := net.ParseCIDR(s)
		if err != nil {
			log.Printf("ERROR: invalid safe ip %q: %v", s, err)
			continue
		}
		safe = append(safe, n)
	}
	return &admin{
		store: NewStore(path),
		safe:  safe,
	}
}

type admin struct {
	store *Store
	safe  []*net.IPNet
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !a.allowed(r) {
		http.Error(w, "not allowed", http.StatusForbidden)
		return
	}
	host := r.Host
	if h := r.Header.Get("X-Host-Override"); h != "" {
		host = h
	}

	pth := r.URL.Path
	if strings.HasPrefix(pth, host+"/") {
		pth = pth[len(host+"/"):]
	}

	var t rest
	switch pth {
	case "":
		t = &collection{name: host, store: a.store}
	default:
		t = &item{name: host + "/" + pth, store: a.store}
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
func (a *admin) allowed(r *http.Request) bool {
	if a != nil {
		// check if the client ip is in the "safe" networks
		rip := r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
		ip := net.ParseIP(strings.Map(cleanIPs, rip))
		for _, s := range a.safe {
			if s.Contains(ip) {
				return true
			}
		}
		// possibly fallback to some sort of authentication scheme
		// TODO
	}
	return false
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
	name  string
	store *Store
}

func (c *collection) get(w http.ResponseWriter, r *http.Request) {
	switch r, err := c.store.List(c.name); err {
	case nil:
		writeJSON(w, r)
	case errNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func (c *collection) put(w http.ResponseWriter, r *http.Request)     {}
func (c *collection) post(w http.ResponseWriter, r *http.Request)    {}
func (c *collection) del(w http.ResponseWriter, r *http.Request)     {}
func (c *collection) patch(w http.ResponseWriter, r *http.Request)   {}
func (c *collection) options(w http.ResponseWriter, r *http.Request) {}
func (c *collection) head(w http.ResponseWriter, r *http.Request)    {}
func (c *collection) def(w http.ResponseWriter, r *http.Request)     {}

type item struct {
	name  string
	store *Store
}

func (i *item) get(w http.ResponseWriter, r *http.Request) {
	switch r, err := i.store.Read(i.name); err {
	case nil:
		writeJSON(w, r)
	case errNotFound:
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
func (i *item) put(w http.ResponseWriter, r *http.Request)     {}
func (i *item) post(w http.ResponseWriter, r *http.Request)    {}
func (i *item) del(w http.ResponseWriter, r *http.Request)     {}
func (i *item) patch(w http.ResponseWriter, r *http.Request)   {}
func (i *item) options(w http.ResponseWriter, r *http.Request) {}
func (i *item) head(w http.ResponseWriter, r *http.Request)    {}
func (i *item) def(w http.ResponseWriter, r *http.Request)     {}

func writeJSON(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "error writing response: %v", err)
	}
}

func cleanIPs(r rune) rune {
	switch r {
	case '[', ']', ' ', '\t', '\r', '\n':
		return -1
	default:
		return r
	}
}
