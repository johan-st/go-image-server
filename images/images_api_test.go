package images_test

import (
	"os"
	"testing"

	img "github.com/johan-st/go-image-server/images"
)

const (
	testFsDir          = "test-fs"
	test_import_source = testFsDir + "/originals"
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

	_, err = os.Stat(originalsDir + "/" + id.String() + ".jpg")
	if err != nil {
		t.Fatal(err)
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
