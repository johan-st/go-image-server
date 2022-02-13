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
	req := httptest.NewRequest("GET", "/1", nil)
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	// fmt.Println(w.Result().Header["Content-Type"])
	is.Equal(w.Result().StatusCode, http.StatusOK)
	is.Equal(w.Result().Header["Content-Type"][0], "image/jpeg")
	// is.Equal(w.Result().Header)
}
