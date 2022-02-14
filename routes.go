package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/johan-st/go-image-server/way"
	"gitlab.com/golang-commonmark/markdown"
)

type server struct {
	l      *log.Logger
	router way.Router
}

// Register handlers for routes
func (srv *server) routes() {
	srv.router.HandleFunc("GET", "/", srv.handleDocs())
	srv.router.HandleFunc("GET", "/favicon.ico", srv.handleFavicon())
	srv.router.HandleFunc("GET", "/:id", srv.handleImg())
	srv.router.HandleFunc("GET", "/:id/:filename", srv.handleImg())

}

// HANDLERS

// handleDocs responds to a request with USAGE.md parsed to html.
// It also inlines some rudimentary css.
func (srv *server) handleDocs() http.HandlerFunc {
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

// HandleImg tages a request and tries to return the requested image by id.
// The id is assumed to be given by the path of the request.
// handleImg also takes query parameter into account to deliver a preprocessed version of the image.
func (srv *server) handleImg() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id_str := way.Param(r.Context(), "id")

		id, err := strconv.Atoi(id_str)
		if err != nil {
			srv.respondError(w, r, fmt.Sprintf("Could not parse image id.\nGOT: %s\nID MUST BE AN INTEGER GREATER THAN ZERO", id_str), http.StatusBadRequest)
			return
		}
		q := r.URL.Query()
		pp, err := parseParameters(q)
		if err != nil {
			srv.respondError(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		srv.serveImage(w, r, id, pp)
	}
}

// handleFavicon serves the favicon.ico.
func (srv *server) handleFavicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/favicon.ico")
	}
}

//  HELPERS

// respondError sends out a respons containing an error. This helper function is meant to be generic enough to serve most needs to communicate errors to the users
func (srv *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	srv.l.Printf("ERROR (%d): %s\n", statusCode, msg)
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "<html><h1>%d</h1><pre>%s</pre></html>", statusCode, msg)
}

func (srv *server) serveImage(w http.ResponseWriter, r *http.Request, id int, ir preprocessingParameters) {
	path, err := pathById(id)
	if err != nil {
		srv.respondError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, path)
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
