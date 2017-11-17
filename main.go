package main

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	errNotFound = errors.New("not found")
	dbFile      = "repo.db"
	listen      = ":19980"
	verbose     = false
	impTmpl     = template.Must(template.New("import").Parse(`<!DOCTYPE html>
<html>
  <head>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8" />
    <meta name="go-import" content="{{ .ImportRoot }} {{ .VCS }} {{ .VCSRoot }}">
    <meta http-equiv="refresh" content="0; url=https://godoc.org/{{ .ImportRoot }}{{ .Suffix }}">
  </head>
  <body>
    Redirecting to docs at <a href="https://godoc.org/{{ .ImportRoot }}{{ .Suffix }}">godoc.org/{{ .ImportRoot }}{{ .Suffix }}</a>...
  </body>
</html>`))
	seed = map[string]Repo{
		"jw4.us/import":   Repo{ImportRoot: "jw4.us/import", VCS: "git", VCSRoot: "https://github.com/johnweldon/import", Suffix: ""},
		"jw4.us/mqd":      Repo{ImportRoot: "jw4.us/mqd", VCS: "git", VCSRoot: "https://github.com/jw4/mqd", Suffix: ""},
		"jw4.us/tunnel":   Repo{ImportRoot: "jw4.us/tunnel", VCS: "git", VCSRoot: "https://github.com/johnweldon/tunnel", Suffix: ""},
		"jw4.us/sortcsv":  Repo{ImportRoot: "jw4.us/sortcsv", VCS: "git", VCSRoot: "https://github.com/johnweldon/sortcsv", Suffix: ""},
		"jw4.us/location": Repo{ImportRoot: "jw4.us/location", VCS: "git", VCSRoot: "https://github.com/johnweldon/location-service", Suffix: ""},
	}
)

func main() {
	if p := os.Getenv("IMPORT_LISTEN_ADDRESS"); p != "" {
		listen = p
	}
	if f := os.Getenv("IMPORT_DB_FILE"); f != "" {
		dbFile = f
	}
	if v := os.Getenv("IMPORT_VERBOSE_LOGGING"); v != "" {
		verbose = true
	}

	mux := http.NewServeMux()
	mux.Handle("/", newImportHandler(dbFile, seed))

	server := http.Server{
		Addr:         listen,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      newLogger(mux, os.Stdout, verbose),
	}

	log.Printf("Using db %s", dbFile)
	log.Printf("Listening on %s", listen)
	log.Fatal(server.ListenAndServe())
}

func newImportHandler(path string, seed map[string]Repo) http.Handler {
	store := NewStore(path)
	if err := store.Initialize(seed); err != nil {
		panic(err)
	}
	return &importer{store: store}
}

type importer struct {
	store *Store
}

func (i *importer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	meta, err := i.store.Get(r.Host, r.URL.Path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := impTmpl.Execute(&buf, meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buf.Bytes())
}
