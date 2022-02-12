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

func (s *server) routes() {
	s.router.HandleFunc("GET", "/", s.handleDocs())
	s.router.HandleFunc("GET", "/:img", s.handleImg())

}

// HANDLERS

func (s *server) handleDocs() http.HandlerFunc {
	s.l.Print("s.handleDocs setup")
	md := markdown.New(markdown.XHTMLOutput(true))

	f, err := ioutil.ReadFile("assets/USAGE.md")
	docs := md.RenderToString(f)
	if err != nil {
		s.l.Fatalf("Could not read docs\n%s", err)
	}
	style, err := ioutil.ReadFile("assets/dark.css")
	if err != nil {
		s.l.Fatalf("Could not read assets/dark.css\n%s", err)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			s.respondError(w, r, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Add("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><body><style>%s</style>%s</body></html>", style, docs)
	}
}
func (s *server) handleImg() http.HandlerFunc {
	s.l.Print("s.handleImg setup")
	return func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "img")
	}
}

// RESPONSE HELPERS

func (s *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "<h1>%d:</h1><pre>%s</pre>", statusCode, msg)
}

// OTHER ESSENTIALS

func newServer(l *log.Logger) *server {
	s := &server{
		l:      l,
		router: *way.NewRouter(),
	}
	s.routes()
	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
	msg := r.Method + " | " + r.URL.Path + " | " + r.RemoteAddr
	s.l.Print(msg)
}
