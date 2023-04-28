package images_test

import (
	"testing"
	"time"
	// img "github.com/johan-st/go-image-server/images"
)

func Test_DreamInteraction(t *testing.T) {

	// defaults, errors if not set properly.
	conf,err := img.Config(img.ConfigParams{
			OriginalsDir: "../test_images",
			CacheDir:     "../test_cache",
			// fallback to defaults if image params could not be used
			FallbackToDefaults: true,
			Defaults: img.ImageDefaults{
				Width:       250,
				Height:      250,
				QualityJpeg: 80,
				QualityPng:  80,
				QualityGif:  170,
				MaxSize:     1*img.Megabytes + 500*img.Kilobytes,
			}})


	// image params overrides defaults if set
	p1 := img.ImageJpeg{
		Id:      1, // required
		Width:   1200,
		Quality: 80,
		MaxSize: 300 * img.Kilobytes,
	}
	p2 := img.ImagePng{
		Id:     2, // required
		Height: 350,
	}
	p3 := img.ImageGif{
		Id:      1, // required
		Width:   50,
		Height:  50,
		Quality: 200,
		MaxSize: 1*img.Megabytes + 500*img.Kilobytes,
	}

	// get images. if not cached, create it
	img1Path, err := img.Get(conf, p1)
	img2Path, err := img.Get(conf, p2)
	img3Path, err := img.Get(conf, p3)

	// error handling: maybe somthing like this to allow inteligent retries and custom error messages?
	bool := img.ErrIsIdNotFound(err)
	bool = img.ErrIsIdNotSet(err)
	bool = img.ErrIsBadParams(err)

	// get all ids
	ids = img.ListIds(conf)

}

// clear cache based on rules
	cacheRules := img.CacheRules{
		MaxAge:       24 * time.Hour,
		MaxCacheSize: 20 * img.Gigabytes,
	}
	err = img.CacheClear(conf, cacheRules)

	// clear one image
	err = img.CacheClear(conf, 2)

	// rebuild all cached
	err = img.CacheRebuild(conf)

	// rebuild one image
	err = img.CacheRebuild(conf, 1)

}
