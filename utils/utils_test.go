package utils_test

import (
	"testing"

	"github.com/johan-st/go-image-server/utils"
)

func TestPath(t *testing.T) {
	str := "test"
	p := utils.Path("test")
	if p != "test" {
		t.Errorf("want: %s, got: %s", str, p)
	}

}
