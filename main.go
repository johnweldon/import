package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	dbFile  = "repo.db"
	listen  = ":19980"
	verbose = false
	public  = "public"
	safeIPs = []string{"127.0.0.0/8", "::1/128"}
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
	if p := os.Getenv("PORT"); p != "" {
		listen = ":" + p
	}
	if f := os.Getenv("IMPORT_DB_FILE"); f != "" {
		dbFile = f
	}
	if v := os.Getenv("IMPORT_VERBOSE_LOGGING"); v != "" {
		verbose = true
		log.Printf("Verbose Logging enabled")
	}
	if p := os.Getenv("IMPORT_PUBLIC_DIR"); p != "" {
		public = p
	}
	if s := os.Getenv("IMPORT_SAFE_IPS"); s != "" {
		safeIPs = append(safeIPs, strings.Split(s, ",")...)
	}
	safeNetworks := getNetworks(safeIPs)

	mux := http.NewServeMux()
	mux.Handle("/_api/", http.StripPrefix("/_api/", newAPIHandler(dbFile, safeNetworks)))
	mux.Handle("/", newImportHandler(dbFile, seed, http.FileServer(http.Dir(public))))

	server := http.Server{
		Addr:         listen,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      newLogger(mux, os.Stdout, verbose),
	}

	log.Printf("Listening on %s", listen)
	log.Printf("Using db %s", dbFile)
	log.Printf("Serving from %q", public)
	log.Printf("Allowing API access from %+v", safeNetworks)
	log.Fatal(server.ListenAndServe())
}
