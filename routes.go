package main

import (
	"errors"
	"fmt"
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
	l      *log.Logger
	ih     *images.ImageHandler
	router way.Router
}

// Register handlers for routes
func (srv *server) routes() {
	srv.router.HandleFunc("GET", "/", srv.handleDocs())
	srv.router.HandleFunc("GET", "/favicon.ico", srv.handleFavicon())
	srv.router.HandleFunc("GET", "/clearcache", srv.handleClearCache())
	srv.router.HandleFunc("GET", "/info", srv.handleInfo())
	srv.router.HandleFunc("GET", "/housekeeping", srv.handleHousekeeping())
	srv.router.HandleFunc("GET", "/:id", srv.handleImg())
	srv.router.HandleFunc("GET", "/:id/:filename", srv.handleImg())

}

// HANDLERS

// handleDocs responds to a request with USAGE.md parsed to html.
// It also inlines some rudimentary css.
func (srv *server) handleDocs() http.HandlerFunc {
	// setup
	l := srv.l.With("handler", "handleDocs")
	t := time.Now
	defer l.Debug("docs rendered", "time", time.Since(t()))

	md := markdown.New(markdown.XHTMLOutput(true))

	f, err := os.ReadFile("docs/USAGE.md")
	if err != nil {
		l.Fatalf("Could not read docs\n%s", err)
	}
	docs := md.RenderToString(f)

	style, err := os.ReadFile("docs/dark.css")
	if err != nil {
		srv.l.Fatalf("Could not read docs/dark.css\n%s", err)
	}

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("handling request", "method", r.Method, "path", r.URL.Path)
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
	l := srv.l.With("handler", "handleImg")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("handling request", "method", r.Method, "path", r.URL.Path, "query", r.URL.Query())
		id_str := way.Param(r.Context(), "id")

		id, err := strconv.Atoi(id_str)
		if err != nil {
			l.Warn("count not parse image id")
			srv.respondError(w, r, fmt.Sprintf("Could not parse image id.\nGOT: %s\nID MUST BE AN INTEGER GREATER THAN ZERO", id_str), http.StatusBadRequest)
			return
		}
		iid := images.ImageId(id)

		q := r.URL.Query()
		imgPar, err := parseImageParameters(q)
		if err != nil {
			l.Warn("could not parse image parameters", "err", err, "query", q)
			srv.respondError(w, r, err.Error(), http.StatusBadRequest)
			return
		}
		path, err := srv.ih.Get(imgPar, iid)
		if err != nil {
			if errors.Is(err, images.ErrIdNotFound{}) {
				l.Warn("image not found", "id", id, "ImageParameters", imgPar, "err", err)
				srv.respondError(w, r, err.Error(), http.StatusNotFound)
				return
			}
			l.Error("could not get image", "id", id, "ImageParameters", imgPar, "err", err)
			srv.respondError(w, r, err.Error(), http.StatusInternalServerError)
			return
		}
		l.Debug("serving image", "id", id, "ImageParameters", imgPar, "path", path)
		http.ServeFile(w, r, path)
	}
}

// handleFavicon serves the favicon.ico.
func (srv *server) handleFavicon() http.HandlerFunc {
	// setup
	l := srv.l.With("handler", "handleFavicon")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("favicon requested")
		http.ServeFile(w, r, "assets/favicon.ico")
	}
}

func (srv *server) handleClearCache() http.HandlerFunc {
	// setup
	l := srv.l.With("handler", "handleClearCache")

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error("not implemented")
		srv.respondError(w, r, "not implemented", http.StatusNotImplemented)
	}
}

func (srv *server) handleInfo() http.HandlerFunc {
	// setup
	l := srv.l.With("handler", "handleInfo")
	return func(w http.ResponseWriter, r *http.Request) {
		// handler
		l.Error("not implemented")
		srv.respondError(w, r, "not implemented", http.StatusNotImplemented)
	}
}

// TODO: implement this properly. Is buggy
func (srv *server) handleHousekeeping() http.HandlerFunc {
	// setup
	l := srv.l.With("handler", "handleHousekeeping")
	return func(w http.ResponseWriter, r *http.Request) {
		l.Error("not implemented")
		srv.respondError(w, r, "not implemented", http.StatusNotImplemented)
	}
}

//  HELPERS

// respondError sends out a respons containing an error. This helper function is meant to be generic enough to serve most needs to communicate errors to the users
func (srv *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	srv.l.Warn("responding with error", "status", statusCode, "message", msg)
	fmt.Fprintf(w, "<html><h1>%d</h1><pre>%s</pre></html>", statusCode, msg)
}

func parseImageParameters(val url.Values) (images.ImageParameters, error) {
	p := images.ImageParameters{}
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
		if v, err := parseImageSize(val.Get("maxsize")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("s") {
		if v, err := parseImageSize(val.Get("s")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	}

	err := errors.Join(errs...)
	return p, err
}

// parseImageFormat parses a string into an images.Format.
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

// parseImageSize parses a string into an images.Size.
// The string can be a number or a string with a optional unit.
// Supported units are: B, KB, MB
// no unit is interpreted as Kilobytes.
func parseImageSize(str string) (images.Size, error) {
	strUp := strings.ToUpper(str)
	// check and trim suffixes
	// try to parse as int
	// return error if not possible
	if strings.Contains(strUp, "MB") {
		if v, err := strconv.Atoi(strings.TrimSuffix(strUp, "MB")); err == nil {
			return images.Size(v * images.Megabyte), nil
		} else {
			return images.Size(0), fmt.Errorf("could not parse image size in Megabytes: %s\n(supported units are: B, KB and MB)", str)
		}
	}

	if strings.Contains(strUp, "KB") {
		if v, err := strconv.Atoi(strings.TrimSuffix(strUp, "KB")); err == nil {
			return images.Size(v * images.Kilobyte), nil
		} else {
			return images.Size(0), fmt.Errorf("could not parse image size in Kilobytes: %s\n(supported units are: B, KB and MB)", str)
		}
	}

	if strings.Contains(strUp, "B") {
		if v, err := strconv.Atoi(strings.TrimSuffix(strUp, "B")); err == nil {
			return images.Size(v), nil
		} else {
			return images.Size(0), fmt.Errorf("could not parse image size in Bytes: %s\n(supported units are: B, KB and MB)", str)
		}

	}

	if v, err := strconv.Atoi(strUp); err == nil {
		return images.Size(v * images.Kilobyte), nil
	} else {
		return images.Size(0), fmt.Errorf("could not parse image size: %s\n(supported units are: B, KB and MB)", str)

	}
}

// OTHER ESSENTIALS

func newServer(l *log.Logger, ih *images.ImageHandler) *server {
	srv := &server{
		l:      l,
		router: *way.NewRouter(),
		ih:     ih,
	}
	srv.routes()
	return srv
}

func (srv *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s := time.Now()
	srv.router.ServeHTTP(w, r)
	srv.l.Info("Request served",
		"method", r.Method,
		"url", r.Host+r.URL.String(),
		"remote", r.RemoteAddr,
		"user-agent", r.UserAgent(),
		"duration", time.Since(s),
	)
}
