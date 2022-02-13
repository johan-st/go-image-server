package main

import "fmt"

func pathById(id int) string {
	path := "originals/" + fmt.Sprint(id) + ".jpg"
	return path
}
