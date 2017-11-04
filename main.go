package main

import (
	"bytes"
	"errors"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", importHandler)

	server := http.Server{
		Addr:         ":19980",
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      mux,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

var (
	errNotFound = errors.New("not found")
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
	imports = map[string]*data{
		"jw4.us/mqd": &data{ImportRoot: "jw4.us/mqd", VCS: "git", VCSRoot: "https://github.com/jw4/mqd", Suffix: ".git"},
	}
)

type data struct {
	ImportRoot string `json:"import_root"`
	VCS        string `json:"vcs"`
	VCSRoot    string `json:"vcs_root"`
	Suffix     string `json:"suffix"`
}

func importHandler(w http.ResponseWriter, r *http.Request) {
	dumpRequest(r)
	defer dumpResponse()

	meta := &data{}
	if err := get(r, meta); err != nil {
		log.Printf("returned 404")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var buf bytes.Buffer
	if err := impTmpl.Execute(&buf, meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("returned 200")
	_, _ = w.Write(buf.Bytes())
}

func get(r *http.Request, meta *data) error {
	for _, path := range getImportPaths(r) {
		if m, ok := imports[path]; ok {
			log.Printf("Import path %q found", path)
			*meta = *m
			return nil
		} else {
			log.Printf("Import path %q not found", path)
		}
	}
	return errNotFound
}

func getImportPaths(r *http.Request) []string {
	var paths []string
	if r != nil {
		for segs := strings.Split(strings.TrimSuffix(r.URL.Path, "/"), "/"); len(segs) > 1; segs = segs[:len(segs)-1] {
			paths = append(paths, r.Host+strings.Join(segs, "/"))
		}
	}
	if len(paths) == 0 {
		paths = []string{""}
	}
	return paths
}

func dumpRequest(r *http.Request) {
	if buf, err := httputil.DumpRequest(r, true); err == nil {
		log.Printf("Request: %q\n===\n%s\n===\n", getImportPaths(r)[0], string(buf))
	}
}

func dumpResponse() {
	log.Printf("Done")
}
