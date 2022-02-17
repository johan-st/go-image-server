package main

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"net/url"
	"os"
	"strconv"
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

// loadImage retunes the image specified by the path
func loadImage(path string) (image.Image, error) {

	file, err := os.Open(path)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	return img, nil
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
