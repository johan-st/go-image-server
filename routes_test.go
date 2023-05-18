package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
	"github.com/johan-st/go-image-server/way"
	"github.com/matryer/is"
)

// it is very cool that we can test the the routes and handlers but it is probably not worth the effort in this case.

const (
	testFsDir          = "test-fs"
	test_import_source = "test-images"
)

func Test_HandleDocs(t *testing.T) {
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
	if !strings.Contains(w.Body.String(), "jst_ImageServer") {
		is.Fail()
	}
}

func Test_HandleImg(t *testing.T) {
	is := is.New(t)

	// arrange
	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	is.NoErr(err)

	defer os.RemoveAll(originalsDir)

	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	is.NoErr(err)

	defer os.RemoveAll(cachePath)

	conf := images.Config{
		OriginalsDir: originalsDir,
		CacheDir:     cachePath,
	}

	ih, err := images.New(conf, nil)
	is.NoErr(err)

	id, err := ih.Add(test_import_source + "/one.jpg")
	is.NoErr(err)
	t.Log(id)

	srv := server{
		l:      log.Default(),
		router: *way.NewRouter(),
		ih:     ih,
	}
	srv.routes()
	w := httptest.NewRecorder()

	// act
	idStr := strconv.Itoa(id)
	srv.ServeHTTP(w, httptest.NewRequest("GET", "/"+idStr, nil))
	t.Log(w.Result())

	// assert
	is.Equal(w.Result().StatusCode, http.StatusOK)
	is.Equal(w.Result().Header["Content-Type"][0], "image/jpeg")
	sizeRes := w.Result().Header["Content-Length"][0]
	sizeResInt, _ := strconv.Atoi(sizeRes)

	s := 10 * images.Kilobyte
	if v, _ := strconv.Atoi(sizeRes); v < s {
		t.Fatalf("Content-Lenghth is too small for test image, size: %s, expected at least %s", images.Size(sizeResInt), images.Size(s))
	}
}
