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

	ih, err := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions,
		images.WithCreateDirs,
	)
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

// BENCHMARKS

func Benchmark_HandleDocs(b *testing.B) {
	l := log.WithPrefix("[http bechmark]")
	l.SetLevel(log.FatalLevel)

	srv := server{
		l:      l,
		router: *way.NewRouter(),
	}

	srv.routes()
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(w, req)
	}
}

func Benchmark_HandleImg_cached(b *testing.B) {
	l := log.Default()
	l.SetLevel(log.FatalLevel)

	// arrange
	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(originalsDir)

	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(cachePath)

	ih, err := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions,
		images.WithCreateDirs,
	)

	if err != nil {
		b.Fatal(err)
	}

	id, err := ih.Add(test_import_source + "/one.jpg")
	if err != nil {
		b.Fatal(err)
	}

	srv := server{
		l:      l,
		router: *way.NewRouter(),
		ih:     ih,
	}

	srv.routes()
	w := httptest.NewRecorder()

	// cache image by calling it once
	idStr := strconv.Itoa(id)
	req := httptest.NewRequest("GET", "/"+idStr, nil)
	srv.ServeHTTP(w, req)

	// act
	for i := 0; i < b.N; i++ {
		srv.ServeHTTP(w, req)
	}
}

func Benchmark_HandleImg_notCached(b *testing.B) {
	l := log.Default()
	l.SetLevel(log.FatalLevel)

	// arrange
	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(originalsDir)

	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(cachePath)

	ih, err := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions,
		images.WithCreateDirs,
		images.WithCacheMaxNum(1),
	)

	if err != nil {
		b.Fatal(err)
	}

	id1, err := ih.Add(test_import_source + "/one.jpg")
	if err != nil {
		b.Fatal(err)
	}
	id2, err := ih.Add(test_import_source + "/two.jpg")
	if err != nil {
		b.Fatal(err)
	}

	srv := server{
		l:      l,
		router: *way.NewRouter(),
		ih:     ih,
	}

	srv.routes()
	w := httptest.NewRecorder()

	idStr1 := strconv.Itoa(id1)
	req1 := httptest.NewRequest("GET", "/"+idStr1, nil)

	idStr2 := strconv.Itoa(id2)
	req2 := httptest.NewRequest("GET", "/"+idStr2, nil)

	// act
	for i := 0; i < b.N; i++ {
		if i%2 == 0 {
			srv.ServeHTTP(w, req1)
			continue
		}

		srv.ServeHTTP(w, req2)
	}
}
