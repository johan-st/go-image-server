package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {

	// catFile, err := os.Open("originals/yoshi.png")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer catFile.Close()

	// imData, imType, err := image.Decode(catFile)
	// if err != nil {
	// 	fmt.Println(err)
	// }

	// fmt.Println(imData)
	// fmt.Println(imType)

	// cat, err := png.Decode(catFile)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println(cat)

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
