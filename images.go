package main

import (
	"fmt"
	"net/url"
	"strconv"
)

// pathById handles translating image id's into paths to the original.
func pathById(id int) (string, error) {
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

	return pp, nil
}
