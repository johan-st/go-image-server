package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/johan-st/go-image-server/images"
)

func main() {

	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s/n", err)
		os.Exit(1)
	}
}

func run() error {
	images.ClearCache()
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lf := log.Ldate | log.Ltime
	l := log.New(os.Stdout, "[go-image-server] ", lf)

	srv := newServer(l)
	mainSrv := &http.Server{
		Addr:              ":" + string(port),
		Handler:           srv,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          srv.l,
	}
	srv.l.Printf("server is up and listening on port %s", port)
	return mainSrv.ListenAndServe()
}
