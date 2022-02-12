package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/johan-st/go-image-server/way"
	"gitlab.com/golang-commonmark/markdown"
)

type server struct {
	l      *log.Logger
	router way.Router
}

func (srv *server) routes() {
	srv.router.HandleFunc("GET", "/", srv.handleDocs())
	srv.router.HandleFunc("GET", "/:img", srv.handleImg())

}

// HANDLERS

func (srv *server) handleDocs() http.HandlerFunc {
	// srv.l.Print("s.handleDocs setup")
	md := markdown.New(markdown.XHTMLOutput(true))

	f, err := ioutil.ReadFile("assets/USAGE.md")
	docs := md.RenderToString(f)
	if err != nil {
		srv.l.Fatalf("Could not read docs\n%s", err)
	}
	style, err := ioutil.ReadFile("assets/dark.css")
	if err != nil {
		srv.l.Fatalf("Could not read assets/dark.css\n%s", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			srv.respondError(w, r, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Add("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><body><style>%s</style>%s</body></html>", style, docs)
	}
}
func (srv *server) handleImg() http.HandlerFunc {
	// srv.l.Print("s.handleImg setup")
	return func(w http.ResponseWriter, r *http.Request) {
		img := way.Param(r.Context(), "img")
		fmt.Fprintf(w, "%s", img)
	}
}

// RESPONSE HELPERS

func (srv *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "<h1>%d:</h1><pre>%s</pre>", statusCode, msg)
}

// OTHER ESSENTIALS

func newServer(l *log.Logger) *server {
	srv := &server{
		l:      l,
		router: *way.NewRouter(),
	}
	srv.routes()
	return srv
}

func (srv *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	srv.router.ServeHTTP(w, r)
	msg := r.Method + " | " + r.URL.Path + " | " + r.RemoteAddr
	srv.l.Print(msg)
}
