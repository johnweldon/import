package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
)

func newAPIHandler(path string) http.Handler { return &admin{store: NewStore(path)} }

type admin struct {
	store *Store
}

func (a *admin) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var t rest
	switch r.URL.Path {
	case "":
		t = &collection{name: r.Host, store: a.store}
	default:
		t = &item{name: r.Host + "/" + r.URL.Path, store: a.store}
	}

	log.Printf("admin ServeHTTP: %T %#v", t, t)

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
		p := (&indexPage{Title: c.name, Items: r}).String()
		fmt.Fprintf(w, "%s", p)
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

func (i *item) get(w http.ResponseWriter, r *http.Request)     {}
func (i *item) put(w http.ResponseWriter, r *http.Request)     {}
func (i *item) post(w http.ResponseWriter, r *http.Request)    {}
func (i *item) del(w http.ResponseWriter, r *http.Request)     {}
func (i *item) patch(w http.ResponseWriter, r *http.Request)   {}
func (i *item) options(w http.ResponseWriter, r *http.Request) {}
func (i *item) head(w http.ResponseWriter, r *http.Request)    {}
func (i *item) def(w http.ResponseWriter, r *http.Request)     {}

type indexPage struct {
	Title string
	Body  string
	Items []Repo
}

func (p *indexPage) String() string {
	t, err := template.New("repos").Parse(indexHTML)
	if err != nil {
		panic(err)
	}
	var buf bytes.Buffer
	err = t.Execute(&buf, p)
	if err != nil {
		panic(err)
	}
	return buf.String()
}

const (
	indexHTML = `<!DOCTYPE html>
<html>
<head>
  <title>Index of {{ .Title }}</title>
</head>
<body>
  <h1>{{ .Title }}</h1>
  {{ .Body }}
  <div class="repos">
    <ul>{{range .Items}}
      <li><a href="https://godoc.org/{{.ImportRoot}}">{{.ImportRoot}}</a></li>{{end}}
    </ul>
  </div>
</body>
</html>`
)
