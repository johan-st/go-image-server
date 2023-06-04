package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	"github.com/johan-st/go-image-server/way"
	"gitlab.com/golang-commonmark/markdown"
)

type server struct {
	errorLogger  *log.Logger // *required
	accessLogger *log.Logger // optional

	conf   confHttp
	ih     *images.ImageHandler
	router way.Router
}

// Register handlers for routes
func (srv *server) routes() {

	// Docs / root
	if srv.conf.Docs {
		srv.router.HandleFunc("GET", "/favicon.ico", srv.handleFavicon())
		srv.router.HandleFunc("GET", "", srv.handleDocs())
	}

	// API
	srv.router.HandleFunc("GET", "/api", srv.handleApiDocs())

	srv.router.HandleFunc("GET", "/api/image", srv.handleApiImageGet())
	srv.router.HandleFunc("DELETE", "/api/image/:id", srv.handleApiImageDelete())
	srv.router.HandleFunc("*", "/api/image/", srv.handleNotAllowed())

	// Admin
	srv.router.HandleFunc("GET", "/admin", srv.handleAdmin())

	// upload
	srv.router.HandleFunc("POST", "/upload", srv.handleUpload())

	// Serve Images
	srv.router.HandleFunc("GET", "/:id/:preset/", srv.handleImgWithPreset())
	srv.router.HandleFunc("GET", "/:id/", srv.handleImg())

	// 404
	srv.router.NotFound = srv.handleNotFound()
}

// HANDLERS

func (srv *server) handleAdmin() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleAdmin")

	// time the handler initialization
	defer func(t time.Time) {
		l.Debug("admin page parsed and ready to be served", "time", time.Since(t))
	}(time.Now())

	// template, err := template.ParseFiles("www/admin.gohtml", "www/index.gohtml")
	template, err := template.ParseFiles("www/index.gohtml")
	if err != nil {
		l.Fatalf("Could not parse admin template\n%s", err)
	}

	type data struct {
		Title   string
		Content string
		Name    string
	}

	d := data{
		Title:   "Admin",
		Content: "This is the admin page",
		Name:    "Mr Admin",
	}
	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		respondTemplate(w, r, http.StatusOK, template, d)
	}
}

// handleDocs responds to a request with USAGE.md parsed to html.
// It also inlines some rudimentary css.
func (srv *server) handleDocs() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleDocs")

	// time the handler initialization
	defer func(t time.Time) {
		l.Debug("docs rendered and ready to be served", "time", time.Since(t))
	}(time.Now())

	md := markdown.New(markdown.XHTMLOutput(true))

	f, err := os.ReadFile("docs/USAGE.md")
	if err != nil {
		l.Fatalf("Could not read docs\n%s", err)
	}
	docs := md.RenderToString(f)

	style, err := os.ReadFile("docs/dark.css")
	if err != nil {
		l.Fatalf("Could not read docs/dark.css\n%s", err)
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path)
		if r.Method != http.MethodGet {
			srv.respondError(w, r, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Add("content-type", "text/html")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "<html><head><title>img.jst.dev | no hassle image hosting</title></head><body><style>%s</style>%s</body></html>", style, docs)
	}
}

// HandleImg tages a request and tries to return the requested image by id.
// The id is assumed to be given by the path of the request.
// handleImg also takes query parameter into account to deliver a preprocessed version of the image.
func (srv *server) handleImg() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleImg")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path, "query", r.URL.Query())
		id_str := way.Param(r.Context(), "id")

		id, err := strconv.Atoi(id_str)
		if err != nil {
			l.Warn("count not parse image id")
			srv.respondError(w, r, fmt.Sprintf("Could not parse image id.\nGOT: %s\nID MUST BE AN INTEGER GREATER THAN ZERO", id_str), http.StatusBadRequest)
			return
		}

		q := r.URL.Query()
		imgPar, err := parseImageParameters(id, q)
		if err != nil {
			l.Warn("could not parse image parameters", "err", err, "query", q)
			srv.respondError(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		srv.respondWithImage(w, r, l, imgPar)
	}
}

func (srv *server) handleImgWithPreset() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleImgWithPreset")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path, "query", query)

		id_str := way.Param(r.Context(), "id")
		id, err := strconv.Atoi(id_str)
		if err != nil {
			l.Warn("count not parse image id")
			srv.respondError(w, r, fmt.Sprintf("Could not parse image id.\nGOT: %s\nID MUST BE AN INTEGER GREATER THAN ZERO", id_str), http.StatusBadRequest)
			return
		}

		presetMaybe := way.Param(r.Context(), "preset")
		p, ok := srv.ih.GetPreset(presetMaybe)
		if !ok {
			l.Debug("could not match preset. ignoring..", "not a preset", presetMaybe)
			imgPar, err := parseImageParameters(id, query)
			if err != nil {
				l.Warn("could not parse image parameters", "err", err, "query", query)
				srv.respondError(w, r, err.Error(), http.StatusBadRequest)
				return
			}
			srv.respondWithImage(w, r, l, imgPar)
			return
		}

		l.Debug("found preset alias",
			"alias", presetMaybe,
			"preset", p.Name)

		imgPar, err := parseImageParametersWithPreset(id, query, p)
		if err != nil {
			l.Warn("could not parse image parameters", "err", err, "query", query)
			srv.respondError(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		srv.respondWithImage(w, r, l, imgPar)
	}
}

// handleFavicon serves the favicon.ico.
func (srv *server) handleFavicon() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleFavicon")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path)
		http.ServeFile(w, r, "assets/favicon.ico")
	}
}

func (srv *server) handleNotFound() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleNotFound")
	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path)
		srv.respondError(w, r, "not found", http.StatusNotFound)
	}
}
func (srv *server) handleNotAllowed() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleNotAllowed")
	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("Method not allowed", "method", r.Method, "path", r.URL.Path)
		respondCode(w, r, http.StatusMethodNotAllowed)
	}
}

// RESPONDERS

func respondTemplate(w http.ResponseWriter, r *http.Request, code int, tmpl *template.Template, data interface{}) {
	w.WriteHeader(code)
	tmpl.Execute(w, data)
}

func respondJson(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func respondCode(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
}

// respondError sends out a respons containing an error. This helper function is meant to be generic enough to serve most needs to communicate errors to the users
func (srv *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "<html><h1>%d</h1><pre>%s</pre></html>", statusCode, msg)
}

//  HELPERS

func (srv *server) respondWithImage(w http.ResponseWriter, r *http.Request, l *log.Logger, imgPar images.ImageParameters) {
	path, err := srv.ih.Get(imgPar)
	if err != nil {
		if errors.Is(err, images.ErrIdNotFound{}) {
			l.Warn("id not found", "id", imgPar.Id, "referer", r.Referer())
			srv.respondError(w, r, fmt.Sprintf("id '%d' was not found", imgPar.Id), http.StatusNotFound)
			return
		}
		l.Error("failed to serve image", "id", imgPar.Id, "ImageParameters", imgPar, "err", err)
		srv.respondError(w, r, err.Error(), http.StatusInternalServerError)
		return
	}
	l.Debug("serving image", "id", imgPar.Id, "ImageParameters", imgPar, "file", path)
	http.ServeFile(w, r, path)
}

func parseImageParameters(id int, val url.Values) (images.ImageParameters, error) {
	p := images.ImageParameters{Id: id}
	errs := []error{}

	if val.Has("width") {
		if v, err := strconv.ParseUint(val.Get("width"), 10, 32); err == nil {
			p.Width = uint(v)
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("w") {
		if v, err := strconv.ParseUint(val.Get("w"), 10, 32); err == nil {
			p.Width = uint(v)
		} else {
			errs = append(errs, err)
		}

	}

	if val.Has("height") {
		if v, err := strconv.ParseUint(val.Get("height"), 10, 32); err == nil {
			p.Height = uint(v)
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("h") {
		if v, err := strconv.ParseUint(val.Get("h"), 10, 32); err == nil {
			p.Height = uint(v)
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("quality") {
		if v, err := strconv.Atoi(val.Get("quality")); err == nil {
			p.Quality = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("q") {
		if v, err := strconv.Atoi(val.Get("q")); err == nil {
			p.Quality = v
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("format") {
		if v, err := parseImageFormat(val.Get("format")); err == nil {
			p.Format = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("f") {
		if v, err := parseImageFormat(val.Get("f")); err == nil {
			p.Format = v
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("maxsize") {
		if v, err := images.ParseSize(val.Get("maxsize")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("s") {
		if v, err := images.ParseSize(val.Get("s")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	}

	err := errors.Join(errs...)
	return p, err
}

func parseImageParametersWithPreset(id int, val url.Values, pre images.ImagePreset) (images.ImageParameters, error) {
	p := images.ImageParameters{
		Id:      id,
		Width:   uint(pre.Width),
		Height:  uint(pre.Height),
		Quality: pre.Quality,
		Format:  pre.Format,
		MaxSize: pre.MaxSize,
	}
	errs := []error{}

	if val.Has("width") {
		if v, err := strconv.ParseUint(val.Get("width"), 10, 32); err == nil {
			p.Width = uint(v)
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("w") {
		if v, err := strconv.ParseUint(val.Get("w"), 10, 32); err == nil {
			p.Width = uint(v)
		} else {
			errs = append(errs, err)
		}

	}

	if val.Has("height") {
		if v, err := strconv.ParseUint(val.Get("height"), 10, 32); err == nil {
			p.Height = uint(v)
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("h") {
		if v, err := strconv.ParseUint(val.Get("h"), 10, 32); err == nil {
			p.Height = uint(v)
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("quality") {
		if v, err := strconv.Atoi(val.Get("quality")); err == nil {
			p.Quality = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("q") {
		if v, err := strconv.Atoi(val.Get("q")); err == nil {
			p.Quality = v
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("format") {
		if v, err := parseImageFormat(val.Get("format")); err == nil {
			p.Format = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("f") {
		if v, err := parseImageFormat(val.Get("f")); err == nil {
			p.Format = v
		} else {
			errs = append(errs, err)
		}
	}

	if val.Has("maxsize") {
		if v, err := images.ParseSize(val.Get("maxsize")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("s") {
		if v, err := images.ParseSize(val.Get("s")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	}

	err := errors.Join(errs...)
	return p, err
}

// parseImageFormat parses a string into an images.Format.
// TODO: cam i return an "ok" bool here instead of an error?
func parseImageFormat(str string) (images.Format, error) {
	strUp := strings.ToUpper(str)
	switch strUp {
	case "JPG":
		return images.Jpeg, nil
	case "JPEG":
		return images.Jpeg, nil
	case "PNG":
		return images.Png, nil
	case "GIF":
		return images.Gif, nil
	default:
		return images.Jpeg, fmt.Errorf("could not parse image format: %s\n(supported formats are: jpg (/jpeg), png and gif)", str)
	}
}

// OTHER ESSENTIALS

func (srv *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if srv.accessLogger == nil {
		srv.router.ServeHTTP(w, r)
		return
	}
	t := time.Now()
	srv.router.ServeHTTP(w, r)
	srv.accessLogger.Print(t.UTC().Local(),
		"method", r.Method,
		"url", r.Host+r.URL.String(),
		"remote", r.RemoteAddr,
		"user-agent", r.UserAgent(),
		"time elapsed", time.Since(t))
}
