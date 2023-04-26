package main

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/johan-st/go-image-server/images"
)

/*
type config struct {
	originalsDir string
	cacheDir     string
}
*/

func main() {
	err := run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run() error {
	err := checkFilePermissions([]string{"originals", "cache"})
	if err != nil {
		return err
	}

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

// check that the images directory exists and is writable. If not, set up needed permissions.
// TODO: folders should be configurable
func checkFilePermissions(paths []string) error {
	for _, path := range paths {
		err := filepath.WalkDir(path, permAtLeast(0700, 0600))
		if err != nil {
			return err
		}
	}
	return nil
}

// Will extend permissions if needed, but will not reduce them.
func permAtLeast(dir os.FileMode, file os.FileMode) fs.WalkDirFunc {
	return func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		i, err := d.Info()
		if err != nil {
			return err
		}
		perm := os.FileMode(0)
		if d.IsDir() {
			perm = dir
		} else {
			perm = file
		}

		if i.Mode().Perm()&perm < perm {
			p := i.Mode().Perm() | perm
			err := os.Chmod(path, p)
			if err != nil {
				return fmt.Errorf("'%s' has insufficient permissions and setting new permission failed. (was: %o, need at least: %o)", path, i.Mode().Perm(), perm)
			}
			fmt.Printf("'%s' had insufficient permissions. Setting permissions to %o. (was: %o, need at least: %o)\n", path, p, i.Mode().Perm(), perm)

		}
		return nil

	}
}
