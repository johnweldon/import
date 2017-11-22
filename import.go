package main

import (
	"bytes"
	"net/http"
)

func newImportHandler(path string, seed map[string]Repo, fallback http.Handler) http.Handler {
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
