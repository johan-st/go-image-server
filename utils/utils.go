package utils

import (
	"path/filepath"
)

func Path(path string) string {
	p := filepath.Clean(path)
	ap, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return ap
}
