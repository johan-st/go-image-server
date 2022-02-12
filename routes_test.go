package main

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/johan-st/go-image-server/way"
	"github.com/matryer/is"
)

func TestHandleDocs(t *testing.T) {
	is := is.New(t)
	srv := server{
		l:      log.Default(),
		router: *way.NewRouter(),
	}
	srv.routes()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	if !strings.Contains(w.Body.String(), "<h1>go-image-server</h1>") {
		is.Fail()
	}
}

func TestHandleImg(t *testing.T) {
	is := is.New(t)
	srv := server{
		l:      log.Default(),
		router: *way.NewRouter(),
	}
	srv.routes()
	req := httptest.NewRequest("GET", "/blobbers", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	is.Equal(w.Code, http.StatusOK)
	if !strings.Contains(w.Body.String(), "blobbers") {
		srv.l.Println(w.Body)
		is.Fail()
	}
}
