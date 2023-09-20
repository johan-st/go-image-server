package main

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/a-h/templ"
	log "github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	components "github.com/johan-st/go-image-server/pages/components"
	"github.com/johan-st/go-image-server/units/size"
	"github.com/johan-st/go-image-server/way"
)

//go:embed pages/assets
var staticFS embed.FS
var (
	darkTheme = components.Theme{
		ColorPrimary:    "#f90",
		ColorSecondary:  "#fa3",
		ColorBackground: "#333",
		ColorText:       "#aaa",
		ColorBorder:     "#666",
		BorderRadius:    "1rem",
	}
	
	metadata = map[string]string{
		"Description": "img.jst.dev is a way for Johan Strand to learn more Go and web development.",
		"Keywords":    "image, hosting",
		"Author":      "Johan Strand",
	}
)


type server struct {
	errorLogger  *log.Logger // *required
	accessLogger *log.Logger // optional

	conf   confHttp
	ih     *images.ImageHandler
	router way.Router

	// TODO: make concurrent safe
	Stats struct {
		StartTime    time.Time
		Requests     int
		Errors       int
		ImagesServed int
	}
}

// Register handlers for routes
func (srv *server) routes() {

	srv.Stats.StartTime = time.Now()

	// Docs / root
	if srv.conf.Docs {
		srv.router.HandleFunc("GET", "", srv.handleDocs())
	}

	// STATIC ASSETS
	srv.router.HandleFunc("GET", "/favicon.ico", srv.handleFavicon())
	srv.router.HandleFunc("GET", "/assets/", srv.handleAssets())

	// API
	srv.router.HandleFunc("GET", "/api/images", srv.handleApiImageGet())
	srv.router.HandleFunc("POST", "/api/images", srv.handleApiImagePost())
	srv.router.HandleFunc("DELETE", "/api/images/:id", srv.handleApiImageDelete())
	srv.router.HandleFunc("*", "/api/", srv.handleNotAllowed())

	// Admin
	srv.router.HandleFunc("GET", "/admin", srv.handleAdminTempl())
	srv.router.HandleFunc("GET", "/admin/:page", srv.handleAdminTempl())
	srv.router.HandleFunc("GET", "/admin/images/:id", srv.handleAdminImage())

	// Serve Images
	srv.router.HandleFunc("GET", "/:id/:preset/", srv.handleImgWithPreset())
	srv.router.HandleFunc("GET", "/:id/", srv.handleImg())

	// 404
	srv.router.NotFound = srv.handleNotFound()
}

// HANDLERS

func (srv *server) handleAdminTempl() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleAdminTempl")

	// get base css styles
	styles, err := os.ReadFile("pages/assets/admin.css")
	if err != nil {
		l.Fatal("Could not read admin.css", "error", err)
	}

	baseStyles,err := components.StyleTag(darkTheme, string(styles))
	if err != nil {
		l.Fatal("Could not create style tag", "error", err)
	}


	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(t time.Time) {
			l.Debug("serving admin page",
				"time", time.Since(t),
				"path", r.URL.Path)
		}(time.Now())

		var content templ.Component
		page := way.Param(r.Context(), "page")

		switch page {
		case "add":
			content = components.AddImage()
		case "images":
			ids, err := srv.ih.Ids()
			if err != nil {
				srv.respondError(w, r, "Could not get image ids", http.StatusInternalServerError)
				l.Fatal("Could not get image ids", "error", err)
			}
			strIds := make([]string, len(ids))
			for i, id := range ids {
				strIds[i] = strconv.Itoa(id)
			}

			content = components.Images(strIds)
		case "info":
			info, err := srv.getServerInfo()
			if err != nil {
				l.Error("Could not get server info", "error", err)
			}
			content = components.Info(info)
		default:
			file,err := os.ReadFile("docs/USAGE.md")
			if err != nil {
				l.Error("Could not read docs/USAGE.md", "error", err)
			}

			content = components.MarkdownFile(file)
		}

		layout := components.Admin("img.jst.dev", metadata, baseStyles, content)

		err = layout.Render(r.Context(), w)
		if err != nil {
			l.Error("Could not render template", "error", err)
			srv.respondError(w, r, "err", http.StatusInternalServerError)
		}
	}
}

func (srv *server) getServerInfo() (components.ServerInfo, error) {
	stat, err := srv.ih.Stat()

	info := components.ServerInfo{
		Uptime:         time.Since(srv.Stats.StartTime).Round(time.Second),
		Requests:       srv.Stats.Requests,
		Errors:         srv.Stats.Errors,
		ImagesServed:   srv.Stats.ImagesServed,
		Originals:      len(stat.Ids),
		OriginalsSize:  stat.SizeOrig,
		CachedNum:      stat.Cache.NumItems,
		CacheCapacity:  stat.Cache.Capacity,
		CacheSize:      stat.Cache.Size,
		CacheHit:       int(stat.Cache.Hit),
		CacheMiss:      int(stat.Cache.Miss),
		CacheEvictions: int(stat.Cache.Evictions),
	}
	if err != nil {
		info.InfoCollectionError = err.Error()
		return info, err
	}
	return info, nil
}

func (srv *server) handleAdminImage() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleAdminImage")

	// get base css styles
	styles, err := os.ReadFile("pages/assets/admin.css")
	if err != nil {
		l.Fatal("Could not read admin.css", "error", err)
	}



	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "path", r.URL.Path)
		id, err := strconv.Atoi(way.Param(r.Context(), "id"))
		if err != nil {
			l.Warn("Could not parse id", "error", err)
			srv.respondError(w, r, "id not found", http.StatusNotFound)
			return
		}

		stat, err := srv.ih.StatId(id)
		if err != nil {
			if errors.Is(err, images.ErrIdNotFound{}) {
				l.Warn("id not found", "id", id, "referer", r.Referer())
				srv.respondError(w, r, fmt.Sprintf("id '%d' was not found", id), http.StatusNotFound)
				return
			}
			l.Error("Could not get stats from imagehandler", "error", err, "id", id)
			srv.respondError(w, r, "internal server error", http.StatusInternalServerError)
			return
		}

		info := components.ImageInfo{
			Id:            strconv.Itoa(id),
			OriginalsSize: stat.OriginalSize.String(),
			CachedNum:     strconv.Itoa(stat.CacheNum),
			CacheSize:     stat.CacheSize.String(),
		}

		content := components.Image(info)
		baseStyles,err := components.StyleTag(darkTheme, string(styles))
		if err != nil {
			l.Fatal("Could not create style tag", "error", err)
		}
		layout := components.Admin("img.jst.dev", metadata, baseStyles, content)

		err = layout.Render(r.Context(),w)
		if err != nil {
			l.Error("Could not render template", "error", err)
			srv.respondError(w, r, "err", http.StatusInternalServerError)
		}
	}
}

func (srv *server) handleAssets() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleAssets")
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "path", r.URL.Path)
		file := strings.TrimPrefix(r.URL.Path, "/assets/")

		if file == "" {
			srv.respondError(w, r, "not found", http.StatusNotFound)
			return
		}

		p, err := staticFS.ReadFile("pages/assets/" + file)
		if err != nil {
			srv.respondError(w, r, "not found", http.StatusNotFound)
			return
		}

		mimeType := mime.TypeByExtension(path.Ext(file))
		l.Debug("serving asset", "file", file, "Content-Type", mimeType)
		w.Header().Add("Content-Type", mimeType)

		w.WriteHeader(http.StatusOK)
		w.Write(p)
	}

}

func (srv *server) handleDocs() http.HandlerFunc {
	var docsPath = "docs/USAGE.md"
	// setup
	l := srv.errorLogger.With("handler", "handleDocs")

	// time the handler initialization
	defer func(t time.Time) {
		l.Debug("docs prepared", "time", time.Since(t))
	}(time.Now())

	f, err := os.ReadFile(docsPath)
	if err != nil {
		l.Fatalf("Could not read docs\n%s", err)
	}


	// get base css styles
	styles, err := os.ReadFile("pages/assets/admin.css")
	if err != nil {
		l.Fatal("Could not read admin.css", "error", err)
	}

	baseStyles,err := components.StyleTag(darkTheme, string(styles))
	if err != nil {
		l.Fatal("Could not create style tag", "error", err)
	}


	mdComponent := components.MarkdownFile(f)
	main := components.Docs("img.jst.dev", metadata, baseStyles, mdComponent)

	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "path", r.URL.Path)
		if r.Method != http.MethodGet {
			srv.respondError(w, r, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		err := main.Render(r.Context(), w)
		if err != nil {
			l.Error("Could not render template", "error", err)
			srv.respondError(w, r, "err", http.StatusInternalServerError)
		}
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
			l.Warn("count not parse image id", "id", id_str)
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
		l.Debug("handling request", "method", r.Method, "path", r.URL.Path, "query", r.URL.Query())

		id_str := way.Param(r.Context(), "id")
		id, err := strconv.Atoi(id_str)
		if err != nil {
			l.Warn("count not parse image id")
			srv.respondError(w, r, fmt.Sprintf("Could not parse image id.\nGOT: %s\nID MUST BE AN INTEGER GREATER THAN ZERO", id_str), http.StatusBadRequest)
			srv.Stats.Errors++
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
				srv.Stats.Errors++
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
		l.Debug("handling request", "path", r.URL.Path)
		http.ServeFile(w, r, "pages/assets/favicon.ico")
	}
}

func (srv *server) handleNotFound() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleNotFound")
	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Debug("handling request", "path", r.URL.Path)
		srv.respondError(w, r, "not found", http.StatusNotFound)
	}
}
func (srv *server) handleNotAllowed() http.HandlerFunc {
	// setup
	l := srv.errorLogger.With("handler", "handleNotAllowed")
	// handler
	return func(w http.ResponseWriter, r *http.Request) {
		l.Info("Method not allowed", "path", r.URL.Path)
		srv.respondCode(w, r, http.StatusMethodNotAllowed)
	}
}

// RESPONDERS

func (srv *server) respondJson(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
	if code != http.StatusOK && code != http.StatusCreated {
		srv.Stats.Errors++
	}
}

func (srv *server) respondCode(w http.ResponseWriter, r *http.Request, code int) {
	w.WriteHeader(code)
	if code != http.StatusOK && code != http.StatusCreated {
		srv.Stats.Errors++
	}
}

// respondError sends out a respons containing an error. This helper function is meant to be generic enough to serve most needs to communicate errors to the users
func (srv *server) respondError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	w.WriteHeader(statusCode)
	fmt.Fprintf(w, "<html><h1>%d</h1><pre>%s</pre></html>", statusCode, msg)
	srv.Stats.Errors++
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
	srv.Stats.ImagesServed++
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
		if v, err := size.Parse(val.Get("maxsize")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("s") {
		if v, err := size.Parse(val.Get("s")); err == nil {
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
		if v, err := size.Parse(val.Get("maxsize")); err == nil {
			p.MaxSize = v
		} else {
			errs = append(errs, err)
		}
	} else if val.Has("s") {
		if v, err := size.Parse(val.Get("s")); err == nil {
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

	srv.Stats.Requests++

	srv.router.ServeHTTP(w, r)
	srv.accessLogger.Print(t.UTC().Local(),
		"method", r.Method,
		"url", r.Host+r.URL.String(),
		"remote", r.RemoteAddr,
		"user-agent", r.UserAgent(),
		"time elapsed", time.Since(t))
}
