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

	ihLogger := l.WithPrefix("[ImageHandler]")
	// ihLogger.SetLevel(log.DebugLevel)

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
	srv.l.Infof("server is up and listening on port %s", port)

	return mainSrv.ListenAndServe()
	// return fmt.Errorf("arbitrary error")
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
		TimeFormat:      time.Layout,
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
	return log.InfoLevel
}

func funcCallerFormater(file string, line int, funcName string) string {
	return fmt.Sprintf("%s:%d %s", trimCaller(file, 1, '/'), line, trimCaller(funcName, 1, '.'))
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
