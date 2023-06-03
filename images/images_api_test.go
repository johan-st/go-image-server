package images_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/charmbracelet/log"
	"github.com/johan-st/go-image-server/images"
)

const (
	testFsDir          = "test-fs"
	test_import_source = testFsDir + "/originals"
	commonExt          = ".jpeg" // this needs to be  the same as in images.go
)

func Test_Add(t *testing.T) {
	t.Parallel()
	// Arange
	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(originalsDir)

	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cachePath)

	ih, err := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions(true),
		images.WithCreateDirs(true),
		images.WithLogger(log.New(os.Stderr).WithPrefix(t.Name())),
		images.WithLogLevel("debug"),
	)
	if err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(test_import_source + "/one.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// act
	id, err := ih.Add(file)

	// assert
	if err != nil {
		t.Fatal(err)
	}
	idStr := strconv.Itoa(id)

	path := originalsDir + "/" + idStr + commonExt
	stat, err := os.Stat(originalsDir + "/" + idStr + commonExt)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Size() == 0 {
		t.Log("path\t", path)
		t.Log("size\t", images.Size(stat.Size()))
		t.Log("msg\t", "file is empty")
		t.Fail()
	}

	dir, err := os.ReadDir(originalsDir)
	if err != nil {
		t.Fatal(err)
	}

	if len(dir) < 1 {
		t.Fatal("originals dir is empty")
	}
}

func Test_Get(t *testing.T) {
	t.Parallel()

	// arange¨
	// l := log.Default().WithPrefix("Test_Get").WithLevel(log.LevelDebug)
	// filesystem
	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(originalsDir)

	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(cachePath)
	// handler
	ih, err := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions(true),
		images.WithCreateDirs(true),
		images.WithLogger(log.New(os.Stderr).WithPrefix(t.Name())),
		images.WithLogLevel("debug"),
	)

	if err != nil {
		t.Fatal(err)
	}
	// add original
	id := addOrig(t, ih, test_import_source+"/one.jpg")

	t.Log("added id\t", id)
	// act
	path, err := ih.Get(images.ImageParameters{
		Id:      id,
		Width:   500,
		Height:  500,
		Format:  images.Jpeg,
		Quality: 10,
		MaxSize: 50 * images.Megabyte,
	})
	t.Logf("got path (%s)", path)

	// assert
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Size() < 1024 {
		t.Log("err\t\t file is to small. Bytes:", stat.Size())
		t.Fail()
	}
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	// check dir not empty
	dir, err := os.ReadDir(cachePath)
	if err != nil {
		t.Fatal(err)
	}

	if len(dir) == 0 {
		t.Fatal("cache dir is empty")
	}
}

// func Test_Remove(t *testing.T) {
// 	// arange

// 	originalsDir, err := os.MkdirTemp(testFsDir, "testAdd-Originals_")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(originalsDir)

// 	cachePath, err := os.MkdirTemp(testFsDir, "testAdd-Cache_")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(cachePath)

// 	conf := images.Config{
// 		OriginalsDir: originalsDir,
// 		CacheDir:     cachePath,
// 	}

// 	ih, err := images.New()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	idKeep, err := ih.Add(test_import_source + "/one.jpg")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	idRem, err := ih.Add(test_import_source + "/two.jpg")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	ih.Get(images.ImageParameters{Width: 10}, idKeep)

// 	// act
// 	err = ih.Remove(idRem)

// 	// assert
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	dir, err := os.ReadDir(originalsDir)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	if len(dir) != 1 {
// 		t.Fatalf("originals dir is not the anticipated length (%d). len=%d", 1, len(dir))
// 	}
// 	_, err = ih.Get(images.ImageParameters{Width: 10}, idRem)
// 	if err == nil {
// 		t.Fatal("file still found")
// 	}
// 	if !errors.Is(err, images.ErrIdNotFound{}) {
// 		t.Fatal(err)
// 	}
// 	// check that the keep file is still there
// 	_, err = ih.Get(images.ImageParameters{Width: 10}, idKeep)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// }

func Test_ListIds(t *testing.T) {
	t.Parallel()
	// arange
	originalsDir, _ := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	defer os.RemoveAll(originalsDir)

	cachePath, _ := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	defer os.RemoveAll(cachePath)

	ih, _ := images.New(
		images.WithOriginalsDir(originalsDir),
		images.WithCacheDir(cachePath),
		images.WithSetPermissions(true),
		images.WithCreateDirs(true),
		images.WithLogger(log.New(os.Stderr).WithPrefix(t.Name())),
		images.WithLogLevel("debug"),
	)

	ids := []int{}
	ids = append(ids, addOrig(t, ih, test_import_source+"/one.jpg"))
	ids = append(ids, addOrig(t, ih, test_import_source+"/two.jpg"))
	ids = append(ids, addOrig(t, ih, test_import_source+"/three.jpg"))

	ihList, err := ih.Ids()
	if err != nil {
		t.Fatal(err)
	}

	if len(ids) != len(ihList) {
		t.Fatalf("ids have a different length, expected %d, got %d", len(ids), len(ihList))
	}

	for _, id := range ihList {
		found := false
		for _, id2 := range ids {
			if id == id2 {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("id %d not found", id)
		}
	}
}

// helper

func addOrig(t *testing.T, ih *images.ImageHandler, path string) int {
	t.Helper()
	// add original
	file, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	id, err := ih.Add(file)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
