package pkg

import (
	"bytes"
	"html/template"
	"net/http"
)

func NewImportHandler(path string, seed map[string]Repo, fallback http.Handler) http.Handler {
	store := NewStore(path)
	if err := store.Initialize(seed); err != nil {
		panic(err)
	}
	return &importer{store: store, fallback: fallback}
}

type importer struct {
	store    *Store
	fallback http.Handler
}

func (i *importer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	meta, err := i.store.Read(r.Host + r.URL.Path)
	if err != nil {
		if i.fallback != nil {
			i.fallback.ServeHTTP(w, r)
		} else {
			http.Error(w, err.Error(), http.StatusNotFound)
		}
		return
	}

	var buf bytes.Buffer
	if err := impTmpl.Execute(&buf, meta); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buf.Bytes())
}

var impTmpl = template.Must(template.New("import").Parse(`<!DOCTYPE html>
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
