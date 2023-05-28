package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	"github.com/johan-st/go-image-server/way"
)

func main() {
	l := newCustomLogger()
	log.SetDefault(l)

	err := run()
	if err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			l.Info("server closed")
			os.Exit(0)
			return
		}
		l.Fatal(err)
	}
}

func run() error {
	// parse flags
	flagConf := flag.String("c", "imageServer_config.yaml", "path to configuration file")
	flagDev := flag.Bool("dev", false, "enable source code debugging")
	// flagLogFile := flag.String("log", "", "path to log file")
	flag.Parse()

	// load configuration
	conf, err := loadConfig(*flagConf)

	// set up logger
	l := log.Default()
	// l.Info("starting server", "version", version, "commit", commit, "build time", buildTime)

	// enable development mode before handling first error. If flag is set
	// i.e. report caller and set log level to debug
	if *flagDev {
		l.SetReportCaller(true)
		l.SetLevel(log.DebugLevel)
		conf.LogLevel = "debug"
	}

	// JUST REDIRECT STDOUT TO A FILE INSTEAD (./goImageServer >> file.log)
	// Log to file
	// if *flagLogFile != "" {
	// 	l.Info("logging to file", "path", *flagLogFile)

	// 	l := log.Default()

	// 	logfile, err := os.OpenFile(*flagLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// 	if err != nil {
	// 		l.Fatal("Could no set up logfile", "error", err)
	// 	}
	// 	defer logfile.Close()

	// 	l.SetOutput(logfile)
	// 	l.SetFormatter(log.TextFormatter)
	// 	// l.SetFormatter(log.LogfmtFormatter)
	// 	// l.SetFormatter(log.JSONFormatter)
	// }
	l.Info("starting server...")

	// handle configuration errors
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			l.Error("configuration file not found. creating example configuration in its place", "path", *flagConf)
			saveErr := saveConfig(defaultConfig(), *flagConf)
			if saveErr != nil {
				l.Error("could not save config", "error", saveErr)
				return fmt.Errorf("config file not found. could not create example config file: %w", saveErr)
			}
			return fmt.Errorf("config was not found. created example config in %s", *flagConf)
		}
		return err
	}

	l.Debug("configuration loaded", "file path", *flagConf)
	err = conf.validate()
	if err != nil {
		return err
	}
	l.Debug("configuration loaded and validated", "file path", *flagConf)

	l.SetLevel(log.ParseLevel(conf.LogLevel))

	if conf.Files.ClearOnStart {
		l.Warn(
			"Clearing folders",
			"originals_dir", conf.Files.DirOriginals,
			"cache_dir", conf.Files.DirCache,
		)
		os.RemoveAll(conf.Files.DirOriginals)
		os.RemoveAll(conf.Files.DirCache)
	}

	imageDefaults, err := toImageDefaults(conf.ImageDefaults)
	if err != nil {
		return err
	}

	imagePresets, err := toImagePresets(conf.ImagePresets, imageDefaults)
	if err != nil {
		return err
	}

	cacheMaxSize, err := images.ParseSize(conf.Cache.MaxSize)
	if err != nil {
		return err
	}

	// create image handler
	ih, err := images.New(
		images.WithLogger(l.WithPrefix("[images]")),
		images.WithLogLevel(conf.LogLevel),

		images.WithCreateDirs(conf.Files.CreateDirs),
		images.WithSetPermissions(conf.Files.SetPerms),

		images.WithOriginalsDir(conf.Files.DirOriginals),
		images.WithCacheDir(conf.Files.DirCache),

		images.WithCacheMaxNum(conf.Cache.Cap),
		images.WithCacheMaxSize(cacheMaxSize),

		images.WithImageDefaults(imageDefaults),
		images.WithImagePresets(imagePresets),
	)
	if err != nil {
		return err
	}

	if conf.Files.PopulateFrom != "" {
		err = addFolder(ih, conf.Files.PopulateFrom)
		if err != nil {
			l.Error("could not populate originals", "error", err)
		}
	}

	if conf.Http.Port == 0 {
		conf.Http.Port = 8000
		l.Info("Port not set in config. Using default port", "port", conf.Http.Port)
	}

	// set up http log
	var al *log.Logger
	if conf.Http.AccessLog != "" {
		l.Info("access log enabled", "path", conf.Http.AccessLog)
		file, err := os.OpenFile(conf.Http.AccessLog, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			l.Error("could not open access log file", "error", err)
		}
		defer file.Close()

		al := log.New(file)
		if path.Ext(conf.Http.AccessLog) == ".json" {
			al.SetFormatter(log.JSONFormatter)
			l.Info("access log format set to json", "path", conf.Http.AccessLog)
		} else {
			al.SetFormatter(log.TextFormatter)
			l.Info("access log format set to text", "path", conf.Http.AccessLog)
		}
	}

	// set up srv
	srv := &server{
		conf:   conf.Http,
		router: *way.NewRouter(),
		ih:     ih,
		l:      al,
	}
	srv.routes()

	// set up server
	mainSrv := &http.Server{
		Addr:              fmt.Sprintf("%s:%d", conf.Http.Host, conf.Http.Port),
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
			l.Debug("signal recieved", "signal", sig)
			if conf.Files.ClearOnExit {
				l.Warn(
					"Removing folders",
					"originals_dir", conf.Files.DirOriginals,
					"cache_dir", conf.Files.DirCache,
				)
				os.RemoveAll(conf.Files.DirOriginals)
				os.RemoveAll(conf.Files.DirCache)
			}
			mainSrv.Shutdown(context.Background())
		}
	}()

	l.Info("server is up and listening", "addr", mainSrv.Addr)
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
	return log.NewWithOptions(os.Stdout, opt)
}

func funcCallerFormater(file string, line int, funcName string) string {
	return fmt.Sprintf("%s:%d %s", trimCaller(file, 2, '/'), line, trimCaller(funcName, 2, '.'))
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

func addFolder(ih *images.ImageHandler, folder string) error {
	dir, err := os.Open(folder)
	if err != nil {
		return err
	}
	defer dir.Close()

	files, err := dir.Readdir(0)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		id, err := ih.Add(folder + "/" + file.Name())
		if err != nil {
			log.Default().Info("failed to add image", "file", file.Name(), "error", err)
		} else {
			log.Default().Debug("added image", "file", file.Name(), "id", id)
		}

	}
	return nil
}
