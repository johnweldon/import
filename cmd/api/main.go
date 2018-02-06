package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"jw4.us/import/pkg"
)

var (
	dbFile  = "repo.db"
	listen  = ":19980"
	verbose = false
	public  = "public"
	safeIPs = []string{"127.0.0.0/8", "::1/128"}
)

func init() {
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
}

func main() {
	safeNetworks := pkg.GetNetworks(safeIPs)

	mux := http.NewServeMux()
	mux.Handle("/_api/", http.StripPrefix("/_api/", pkg.NewAPIHandler(dbFile, safeNetworks)))
	mux.Handle("/", pkg.NewImportHandler(dbFile, nil, http.FileServer(http.Dir(public))))

	server := http.Server{
		Addr:         listen,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		Handler:      pkg.NewLogger(mux, os.Stdout, verbose),
	}

	log.Printf("Listening on %s", listen)
	log.Printf("Using db %s", dbFile)
	log.Printf("Serving from %q", public)
	log.Printf("Allowing API access from %+v", safeNetworks)
	log.Fatal(server.ListenAndServe())
}
