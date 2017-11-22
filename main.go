package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	dbFile  = "repo.db"
	listen  = ":19980"
	verbose = false
	public  = "public"
	impTmpl = template.Must(template.New("import").Parse(`<!DOCTYPE html>
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
	if p := os.Getenv("IMPORT_PUBLIC_DIR"); p != "" {
		public = p
	}

	mux := http.NewServeMux()
	mux.Handle("/_api/", http.StripPrefix("/_api/", newAPIHandler(dbFile)))
	mux.Handle("/", newImportHandler(dbFile, seed, http.FileServer(http.Dir(public))))

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
