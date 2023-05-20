package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
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
	conf, err := loadConfig("config.yaml")
	if err != nil {
		return err
	}
	err = conf.validate()
	if err != nil {
		return err
	}

	saveConfig(conf, "runningConfig.yaml")

	ih, err := images.New(
		images.WithLogger(ihLogger),
		images.WithCreateDirs,
		images.WithSetPermissions,
		images.WithLogLevel("debug"),
	)
	if err != nil {
		return err
	}

	addFolder(ih, "test-images")

	port := conf.Server.Port
	if port == 0 {
		l.Fatal("port not set in config")
	}

	srv := newServer(l.WithPrefix("[http]"), ih)
	mainSrv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", conf.Server.Host, port),
		Handler:           srv,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	// graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		for sig := range signalChan {
			l.Warn("shutting down server...", "signal", sig)
			if l.GetLevel() == log.DebugLevel {
				l.Debug("removing test images")
				os.RemoveAll(conf.Handler.Paths.Originals)
				os.RemoveAll(conf.Handler.Paths.Cache)
			}
			mainSrv.Shutdown(context.Background())

		}
	}()

	srv.l.Infof("server is up and listening on %s", mainSrv.Addr)
	return mainSrv.ListenAndServe()
}

// LOGGER STUFF

// set up logger
func newCustomLogger() *log.Logger {
	opt := log.Options{
		Prefix:          "[main]",
		Level:           log.InfoLevel,
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

func addFolder(ih *images.ImageHandler, folder string) {
	l := log.Default()
	dir, err := os.Open(folder)
	if err != nil {

		l.Error("could not open folder", "error", err)
		return
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		l.Error("could not read folder", "error", err)
		return
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ih.Add(folder + "/" + file.Name())
	}
}
