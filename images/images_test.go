package images

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/url"
	"os"
	"reflect"
	"testing"
)

func Test_pathById(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"one", args{1}, "originals/1.jpg", false},
		{"one", args{2}, "originals/2.jpg", false},
		{"one", args{99}, "originals/99.jpg", false}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := originalPathById(tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("pathById() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pathById() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseParameters(t *testing.T) {
	type args struct {
		v url.Values
	}

	tests := []struct {
		name    string
		args    args
		want    PreprocessingParameters
		wantErr bool
	}{
		{
			"no parameters",
			args{parseQuery("")},
			PreprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"quality set",
			args{parseQuery("q=50")},
			PreprocessingParameters{_type: "jpeg", quality: 50, width: 0, height: 0},
			false,
		}, {
			"handles mixed parameters",
			args{parseQuery("q=100&w=900&h=450")},
			PreprocessingParameters{_type: "jpeg", quality: 100, width: 900, height: 450},
			false,
		}, {
			"q=100 should succeed",
			args{parseQuery("q=100")},
			PreprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"type jpeg should succeed",
			args{parseQuery("t=jpeg")},
			PreprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"type png should succeed",
			args{parseQuery("t=png")},
			PreprocessingParameters{_type: "png", quality: 100, width: 0, height: 0},
			false,
		}, {
			"t=gif should succeed",
			args{parseQuery("t=gif")},
			PreprocessingParameters{_type: "gif", quality: 256, width: 0, height: 0},
			false,
		}, {
			"width and height set",
			args{parseQuery("w=50&h=500")},
			PreprocessingParameters{quality: 100, width: 50, height: 500, _type: "jpeg"},
			false,
		}, {
			"type jpg should fail",
			args{parseQuery("w=50&h=500&t=jpg")},
			PreprocessingParameters{},
			true,
		}, {
			"type vim should fail",
			args{parseQuery("w=50&t=vim&h=500")},
			PreprocessingParameters{},
			true,
		}, {
			"q=-1 should fail",
			args{parseQuery("q=-1")},
			PreprocessingParameters{},
			true,
		}, {
			"q=abc should fail",
			args{parseQuery("q=abc")},
			PreprocessingParameters{},
			true,
		}, {
			"q=101 should fail",
			args{parseQuery("q=101")},
			PreprocessingParameters{},
			true,
		}, {
			"width of 0 should fail",
			args{parseQuery("w=0")},
			PreprocessingParameters{},
			true,
		}, {
			"height of 0 should fail",
			args{parseQuery("h=0")},
			PreprocessingParameters{},
			true,
		}, {
			"height fails upon recieving string",
			args{parseQuery("h=full")},
			PreprocessingParameters{},
			true,
		}, {
			"width fails upon recieving string",
			args{parseQuery("w=full")},
			PreprocessingParameters{},
			true,
		}, {
			"Values given are very large",
			args{parseQuery("w=999999999999999999&q=100&h=990000")},
			PreprocessingParameters{quality: 100, width: 999999999999999999, height: 990000, _type: "jpeg"},
			false,
		}, {
			"Values given are TOO large",
			args{parseQuery("w=99999999999999999999&q=100&h=990000")},
			PreprocessingParameters{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseParameters(tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseParameters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseParameters() = %v, want %v", got, tt.want)
			}
		})
	}
}
func parseQuery(q string) url.Values {
	v, err := url.ParseQuery(q)
	if err != nil {
		panic(err)
	}
	return v
}

func Test_getCachePath(t *testing.T) {
	type args struct {
		id int
		pp PreprocessingParameters
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"id 1 and no parameters", args{id: 1, pp: PreprocessingParameters{quality: 100, width: 0, height: 0, _type: "jpeg"}}, "cache/1-w0-h0-q100.jpeg"},
		{"id 2 and parameters", args{id: 2, pp: PreprocessingParameters{quality: 95, width: 900, height: 600, _type: "gif"}}, "cache/2-w900-h600-q95.gif"},
		{"id 3 and parameters", args{id: 3, pp: PreprocessingParameters{quality: 30, width: 300, height: 200, _type: "png"}}, "cache/3-w300-h200-q30.png"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetCachePath(tt.args.id, tt.args.pp); got != tt.want {
				t.Errorf("GetCachePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
func Setup_fileExists(path string) {
	file, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	defer file.Close()

}
func Taredown_fileExists() {
	ClearCache()
}
func Test_fileExists(t *testing.T) {
	path := "./cache/3-w300-h200-q30.png"
	faultyPath := "./cache/3-w300-h200-q30.jpeg"
	Setup_fileExists(path)
	defer Taredown_fileExists()

	type args struct {
		cachePath string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"path to file does not exist", args{cachePath: faultyPath}, false},
		{"path to file does exist", args{cachePath: path}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FileExists(tt.args.cachePath); got != tt.want {
				t.Errorf("fileExists() = %v, want %v", got, tt.want)
			}
		})
	}
}
