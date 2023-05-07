package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
)

func main() {
	l := newCustomLogger()

	err := run(l)
	if err != nil {
		l.Fatal(err)
	}
}

func run(l *log.Logger) error {

	// image handler might not need a logger
	// should return errors and let the caller decide how to handle and log them
	ihLogger := l.WithPrefix("[ImageHandler]")
	// ihLogger.SetLevel(log.DebugLevel)
	// DEBUG: clears cache folder on boot. Not intended behaviour
	os.RemoveAll("img")

	ih, err := images.New(images.Config{
		OriginalsDir: "img/originals",
		CacheDir:     "img/cache",
		CreateDirs:   true,
		SetPerms:     true,
	}, ihLogger)
	if err != nil {
		return err
	}

	port := os.Getenv("PORT")
	if port == "" {
		l.Debug("enviornment variable PORT was not specified", "fallback", 8080)
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

	// DEBUG: add some images
	for _, img := range []string{"one.jpg", "two.jpg", "three.jpg", "four.jpg", "five.jpg", "six.png"} {
		_, err := ih.Add("test-images/" + img)
		if err != nil {
			srv.l.Warn("could not add image", "error", err)
		}
	}
	ids, err := ih.ListIds()
	if err != nil {
		srv.l.Error("could not list images", "error", err)
	}
	idsStr := ""
	for i, id := range ids {
		if i > 0 {
			idsStr += ", "
		}

		idsStr += id.String()
	}

	srv.l.Info("available images:", "images", idsStr)
	// DEBUG:end

	srv.l.Infof("server is up and listening on port %s", port)
	return mainSrv.ListenAndServe()
}

// LOGGER STUFF

// set up logger
func newCustomLogger() *log.Logger {
	opt := log.Options{
		Prefix:          "[main]",
		Level:           envLogLevel(),
		ReportCaller:    false,
		CallerFormatter: funcCallerFormater,
		ReportTimestamp: true,
		TimeFormat:      "",
		Formatter:       log.TextFormatter,
		Fields:          []interface{}{},
	}
	l := log.NewWithOptions(os.Stderr, opt)

	if l.GetLevel() == log.DebugLevel {
		l.SetReportCaller(true)
	}

	// logfile, err := os.OpenFile("image-server.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// l.Error("Could no set up logfile", "error", err)
	// }
	// defer logfile.Close()
	// l.SetOutput(logfile)
	// l.SetFormatter(log.JSONFormatter)
	return l
}

// NOTE: is duplicated in imageHandler
func envLogLevel() log.Level {

	switch os.Getenv("LOG") {
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
	return log.InfoLevel
}

func funcCallerFormater(file string, line int, funcName string) string {
	return fmt.Sprintf(" %s:%d %s ", trimCaller(file, 1, '/'), line, trimCaller(funcName, 1, '.'))
}

// Cleanup a path by returning the last n segments of the path only.
func trimCaller(path string, n int, sep byte) string {
	// lovely borrowed from zap
	// nb. To make sure we trim the path correctly on Windows too, we
	// counter-intuitively need to use '/' and *not* os.PathSeparator here,
	// because the path given originates from Go stdlib, specifically
	// runtime.Caller() which (as of Mar/17) returns forward slashes even on
	// Windows.
	//
	// See https://github.com/golang/go/issues/3335
	// and https://github.com/golang/go/issues/18151
	//
	// for discussion on the issue on Go side.

	// Return the full path if n is 0.
	if n <= 0 {
		return path
	}

	// Find the last separator.
	idx := strings.LastIndexByte(path, sep)
	if idx == -1 {
		return path
	}

	for i := 0; i < n-1; i++ {
		// Find the penultimate separator.
		idx = strings.LastIndexByte(path[:idx], sep)
		if idx == -1 {
			return path
		}
	}

	return path[idx+1:]
}
