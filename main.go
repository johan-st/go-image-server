package main

import (
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
	img "github.com/johan-st/go-image-server/images"
)

func main() {
	l := log.New(os.Stderr)
	l.SetLevel(log.InfoLevel)

	err := run(l)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func run(l *log.Logger) error {
	ihLogger := l.WithPrefix("[ImageHandler]")
	// ihLogger.SetLevel(log.DebugLevel)

	ih, err := img.New(img.Config{
		OriginalsDir: "img/originals",
		CacheDir:     "img/cache",
		CreateDirs:   true,
		SetPerms:     true,
	}, ihLogger)
	if err != nil {
		return err
	}
	ih.CacheClear()
	ih.CacheHouseKeeping()
	ih.ListIds()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := newServer(l.WithPrefix("[http]"), ih)
	mainSrv := &http.Server{
		Addr:              ":" + string(port),
		Handler:           srv,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	srv.l.Printf("server is up and listening on port %s", port)
	return mainSrv.ListenAndServe()
}

// NOTE: is duplicated in imageHandler
func logLevel() log.Level {

	switch os.Getenv("LOG_LEVEL") {
	case "DEBUG":
		return log.DebugLevel
	case "INFO":
		return log.InfoLevel
	case "WARN":
		return log.WarnLevel
	case "ERROR":
		return log.ErrorLevel
	case "FATAL":
		return log.FatalLevel
	}
	return log.ErrorLevel
}
