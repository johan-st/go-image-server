package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s/n", err)
		os.Exit(1)
	}
}

func run() error {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lf := log.Ldate | log.Ltime
	l := log.New(os.Stdout, "[go-image-server] ", lf)

	s := newServer(l)
	mainSrv := &http.Server{
		Addr:              ":" + string(port),
		Handler:           s,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       60 * time.Second,
		ErrorLog:          s.l,
	}
	s.l.Printf("server is up and listening on port %s", port)
	return mainSrv.ListenAndServe()
}

func newServer(l *log.Logger) *server {
	s := &server{
		l:      l,
		router: *http.NewServeMux(),
	}
	s.routes()
	return s
}
