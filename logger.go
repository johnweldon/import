package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
)

func newLogger(s *http.ServeMux, w io.Writer, verbose bool) http.Handler {
	ll := log.New(w, "[http] ", log.LstdFlags|log.LUTC)
	return &logger{mux: s, verbose: verbose, l: ll, o: w}
}

type logger struct {
	mux     *http.ServeMux
	verbose bool
	o       io.Writer
	l       *log.Logger
}

func (l *logger) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l.requestLogger(r)
	rw, logResponse := l.responseLogger(r, w)
	defer logResponse()

	l.mux.ServeHTTP(rw, r)
}

func (l *logger) requestLogger(r *http.Request) {
	l.l.Printf("(client %s) %s %s %s [%s]", getIP(r), r.Host, r.Method, r.URL.Path, r.UserAgent())
	if l.verbose {
		b, err := httputil.DumpRequest(r, true)
		if err != nil {
			l.l.Printf("error dumping request: %v", err)
			return
		}
		fmt.Fprint(l.o, "|REQUEST|\n")
		if _, err = l.o.Write(b); err != nil {
			l.l.Printf("error writing dumped request: %v", err)
			return
		}
	}
}

func (l *logger) responseLogger(r *http.Request, w http.ResponseWriter) (http.ResponseWriter, func()) {
	rr := httptest.NewRecorder()
	return rr, func() {
		for k, v := range rr.HeaderMap {
			w.Header()[k] = v
		}
		w.WriteHeader(rr.Code)

		out := []io.Writer{w}

		if l.verbose {
			fmt.Fprint(l.o, "|RESPONSE|\n")
			if err := rr.HeaderMap.Write(l.o); err != nil {
				l.l.Printf("error dumping response headers: %v", err)
			}
			out = append(out, l.o)
		}

		if _, err := rr.Body.WriteTo(io.MultiWriter(out...)); err != nil {
			l.l.Printf("error sending response: %v", err)
		}

		if l.verbose {
			fmt.Fprint(l.o, "\n")
		}

		l.l.Printf("(client %s) %d %s", getIP(r), rr.Code, http.StatusText(rr.Code))
	}
}

func getIP(r *http.Request) string {
	vars := r.URL.Query()
	ip := vars.Get("ip")
	if ip == "" {
		ip = r.Header.Get("X-Forwarded-For")
		if ip == "" {
			ip = r.RemoteAddr[:strings.LastIndex(r.RemoteAddr, ":")]
		}
	}
	return ip
}
