package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

func pathById(id int) (string, error) {
	if id <= 0 {
		return "", errors.New("id must be greter than 0")
	}
	path := "originals/" + fmt.Sprint(id) + ".jpg"
	return path, nil
}

type preprocessingParameters struct {
	quality int
	width   int
	height  int
}

func parseParameters(v url.Values) (preprocessingParameters, error) {
	pp := preprocessingParameters{}

	quality_str := v.Get("q")
	if quality_str != "" {
		quality, err := strconv.Atoi(quality_str)
		if err != nil {
			return preprocessingParameters{}, err
		}
		if quality < 1 || quality > 100 {
			return preprocessingParameters{}, errors.New("parameter q (quality) must be greater than 0 (zero) and less or equal to 100")
		}
		pp.quality = quality
	}

	width_str := v.Get("w")
	if width_str != "" {
		width, err := strconv.Atoi(width_str)
		if err != nil {
			return preprocessingParameters{}, err
		}
		if width < 1 {
			return preprocessingParameters{}, errors.New("parameter w (width) must be greater than 0 (zero)")
		}
		pp.width = width
	}

	height_str := v.Get("h")
	if height_str != "" {
		height, err := strconv.Atoi(height_str)
		if err != nil {
			return preprocessingParameters{}, err
		}
		if height < 1 {
			return preprocessingParameters{}, errors.New("parameter h (height) must be greater than 0 (zero)")
		}
		pp.height = height
	}
	// fmt.Printf("_____________________\n%+v\n____________________\n", v)
	// fmt.Printf("_____________________\n%+v\n____________________\n", pp)
	// log.Fatalf("_____________________\n%+v\n____________________\n", id_str)

	return pp, nil
}
