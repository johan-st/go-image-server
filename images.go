package main

import (
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"net/url"
	"os"
	"strconv"

	"github.com/nfnt/resize"
)

// originalPathById handles translating image id's into paths to the original.
func originalPathById(id int) (string, error) {
	if id <= 0 {
		return "", fmt.Errorf("image id was malfigured\nGOT: %d\nEXPECTED an integer greater than 0 (zero)", id)
	}
	path := "originals/" + fmt.Sprint(id) + ".jpg"
	return path, nil
}

type preprocessingParameters struct {
	quality int
	width   int
	height  int
	_type   string
}

// Constructs the path where the cached image should be saved.
func getCachePath(id int, pp preprocessingParameters) string {
	cName := fmt.Sprintf("%d-w%d-h%d-q%d.%s", id, pp.width, pp.height, pp.quality, pp._type)
	cPath := fmt.Sprintf("cache/%s", cName)
	return cPath
}

func fileExists(cachePath string) bool {
	_, err := os.Open(cachePath)
	return err == nil
}

// loadImage retunes the image specified by the path
func loadImage(path string) (image.Image, error) {

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// cacheImage processes an image (by id) and caches the resuts. Returns cachedPath on success
func processAndCache(id int, pp preprocessingParameters) (string, error) {
	oPath, err := originalPathById(id)
	if err != nil {
		return "", err
	}
	oImg, err := loadImage(oPath)
	if err != nil {
		return "", err
	}
	cPath := getCachePath(id, pp)
	file, err := os.Create(cPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	img := resize.Resize(uint(pp.width), uint(pp.height), oImg, resize.Lanczos3)
	if pp._type == "jpeg" {
		opt := &jpeg.Options{Quality: pp.quality}
		err = jpeg.Encode(file, img, opt)
		if err != nil {
			return "", err
		}
	} else if pp._type == "png" {
		err = png.Encode(file, img)
		if err != nil {
			return "", err
		}
	} else if pp._type == "gif" {
		nc := (pp.quality * 256) / 100
		opt := &gif.Options{NumColors: nc}
		err = gif.Encode(file, img, opt)
		if err != nil {
			return "", err
		}
	}
	return cPath, nil
}

// parseParameters parses url.Values into the data the preprocessor needs to adapt the image to the users request.
func parseParameters(v url.Values) (preprocessingParameters, error) {
	pp := preprocessingParameters{}

	quality_str := v.Get("q")
	if quality_str != "" {
		quality, err := strconv.Atoi(quality_str)
		if err != nil {
			return preprocessingParameters{}, fmt.Errorf("parameter h (quality) could not be parsed\nGOT: %s\nEXPECTED an integer\nINTERNAL ERROR: %s", quality_str, err)
		}
		if quality < 1 || quality > 100 {
			return preprocessingParameters{}, fmt.Errorf("parameter q (quality) out of bounds\nGOT: %d\nEXPECTED q to be greater than 0 (zero) and less or equal to 100", quality)
		}
		pp.quality = quality
	} else {
		pp.quality = 100
	}

	width_str := v.Get("w")
	if width_str != "" {
		width, err := strconv.Atoi(width_str)
		if err != nil {
			return preprocessingParameters{}, fmt.Errorf("parameter h (width) could not be parsed\nGOT: %s\nEXPECTED an integer\nINTERNAL ERROR: %s", width_str, err)
		}
		if width < 1 {
			return preprocessingParameters{}, fmt.Errorf("parameter w (width) out of bounds\nGOT: %d\nEXPECTED w to be greater than 0 (zero)", width)
		}
		pp.width = width
	}

	height_str := v.Get("h")
	if height_str != "" {
		height, err := strconv.Atoi(height_str)
		if err != nil {
			return preprocessingParameters{}, fmt.Errorf("parameter h (height) could not be parsed\nGOT: %s\nEXPECTED an integer\nINTERNAL ERROR: %s", height_str, err)
		}
		if height < 1 {
			return preprocessingParameters{}, fmt.Errorf("parameter h (height) out of bounds\nGOT: %d\nEXPECTED h to be be greater than 0 (zero)", height)
		}
		pp.height = height
	}
	type_str := v.Get("t")
	if type_str != "" {
		if type_str != "jpeg" && type_str != "png" && type_str != "gif" {
			return preprocessingParameters{}, fmt.Errorf("parameter t (type) could not be parsed\nGOT: %s\nEXPECTED an \"jpeg\", \"gif\" or \"png\"", height_str)
		}
		pp._type = type_str
	} else {
		pp._type = "jpeg"
	}
	return pp, nil
}

func clearCache() {
	err := os.RemoveAll("./cache")
	if err != nil {
		fmt.Println(err)
	}
	err = os.Mkdir("./cache", 0666)
	if err != nil {
		fmt.Println(err)
	}
}
