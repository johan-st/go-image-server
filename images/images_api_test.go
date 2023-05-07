package images_test

import (
	"errors"
	"os"
	"testing"

	img "github.com/johan-st/go-image-server/images"
)

const (
	testFsDir          = "test-fs"
	test_import_source = testFsDir + "/originals"
	commonExt          = ".jstimg" // this needs to be  the same as in images.go
)

func Test_Add(t *testing.T) {
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

	conf := img.Config{
		OriginalsDir: originalsDir,
		CacheDir:     cachePath,
	}

	ih, err := img.New(conf, nil)
	if err != nil {
		t.Fatal(err)
	}

	// act
	id, err := ih.Add(test_import_source + "/one.jpg")

	// assert
	if err != nil {
		t.Fatal(err)
	}

	_, err = os.Stat(originalsDir + "/" + id.String() + commonExt)
	if err != nil {
		t.Fatal(err)
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

	// arange
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

	conf := img.Config{
		OriginalsDir: originalsDir,
		CacheDir:     cachePath,
	}

	ih, err := img.New(conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	id, err := ih.Add(test_import_source + "/two.jpg")
	if err != nil {
		t.Fatal(err)
	}

	// act
	path, err := ih.Get(img.ImageParameters{}, id)

	// assert
	if err != nil {
		t.Fatal(err)
	}
	stat, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if stat.Size() == 0 {
		t.Fatal("file is empty")
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

func Test_Remove(t *testing.T) {
	// arange

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

	conf := img.Config{
		OriginalsDir: originalsDir,
		CacheDir:     cachePath,
	}

	ih, err := img.New(conf, nil)
	if err != nil {
		t.Fatal(err)
	}
	idKeep, err := ih.Add(test_import_source + "/one.jpg")
	if err != nil {
		t.Fatal(err)
	}
	idRem, err := ih.Add(test_import_source + "/two.jpg")
	if err != nil {
		t.Fatal(err)
	}
	ih.Get(img.ImageParameters{Width: 10}, idKeep)

	// act
	err = ih.Remove(idRem)

	// assert
	if err != nil {
		t.Fatal(err)
	}

	dir, err := os.ReadDir(originalsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(dir) != 1 {
		t.Fatalf("originals dir is not the anticipated length (%d). len=%d", 1, len(dir))
	}
	_, err = ih.Get(img.ImageParameters{Width: 10}, idRem)
	if err == nil {
		t.Fatal("file still found")
	}
	if !errors.Is(err, img.ErrIdNotFound{}) {
		t.Fatal(err)
	}
	// check that the keep file is still there
	_, err = ih.Get(img.ImageParameters{Width: 10}, idKeep)
	if err != nil {
		t.Fatal(err)
	}

}

func Test_ListIds(t *testing.T) {
	// arange
	originalsDir, _ := os.MkdirTemp(testFsDir, "testAdd-Originals_")
	defer os.RemoveAll(originalsDir)

	cachePath, _ := os.MkdirTemp(testFsDir, "testAdd-Cache_")
	defer os.RemoveAll(cachePath)

	ih, _ := img.New(
		img.Config{
			OriginalsDir: originalsDir,
			CacheDir:     cachePath,
		}, nil)

	ids := []img.ImageId{}
	id, _ := ih.Add(test_import_source + "/one.jpg")
	ids = append(ids, id)
	id, _ = ih.Add(test_import_source + "/two.jpg")
	ids = append(ids, id)
	id, _ = ih.Add(test_import_source + "/three.jpg")
	ids = append(ids, id)

	ihList, err := ih.ListIds()
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
			t.Fatalf("id %s not found", id)
		}
	}
}

// func Test_FullInteraction(t *testing.T) {
// 	t.FailNow()

// 	// arange
// 	originalsDir, err := os.MkdirTemp(testFsDir, "testAPI_originals_")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(originalsDir)
// 	cachePath, err := os.MkdirTemp(testFsDir, "testAPI_cache_")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(cachePath)

// 	conf := img.Config{
// 		OriginalsDir: originalsDir,
// 		CacheDir:     cachePath,
// 		DefaultParams: img.ImageParameters{
// 			Format:  img.Jpeg,
// 			Width:   900,
// 			Height:  600,
// 			Quality: 80,
// 			MaxSize: 1*img.Megabytes + 500*img.Kilobytes,
// 		},
// 		CreateDirs: true,
// 		SetPerms:   true}

// 	ih, err := img.New(conf, nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// act
// 	p1 := img.ImageParameters{}
// 	p2 := img.ImageParameters{
// 		Format:  img.Gif,
// 		Quality: 64,
// 	}
// 	p3 := img.ImageParameters{
// 		Format:  img.Png,
// 		Width:   100,
// 		Height:  400,
// 		Quality: 90,
// 		MaxSize: img.Infinite,
// 	}

// 	// get images. if not cached, create it
// 	img1Path, err := ih.Get(p1, 1)
// 	img2Path, err := ih.Get(p2, 4)
// 	img3Path, err := ih.Get(p3, 5)

// 	if errors.Is(err, img.ErrIdNotFound{}) {
// 		t.Fatal(err)
// 	}

// 	fmt.Println(img1Path)
// 	fmt.Println(img2Path)
// 	fmt.Println(img3Path)

// 	// get all ids
// 	ids, err := ih.ListIds()
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	fmt.Println(ids)

// 	// clear cache based on rules
// 	err = ih.CacheClear()

// 	// clear one image
// 	err = ih.CacheClear()
// }
