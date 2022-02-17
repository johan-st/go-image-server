package main

import (
	"net/url"
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
		want    preprocessingParameters
		wantErr bool
	}{
		{
			"no parameters",
			args{parseQuery("")},
			preprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"quality set",
			args{parseQuery("q=50")},
			preprocessingParameters{_type: "jpeg", quality: 50, width: 0, height: 0},
			false,
		}, {
			"handles mixed parameters",
			args{parseQuery("q=100&w=900&h=450")},
			preprocessingParameters{_type: "jpeg", quality: 100, width: 900, height: 450},
			false,
		}, {
			"q=100 should succeed",
			args{parseQuery("q=100")},
			preprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"type jpeg should succeed",
			args{parseQuery("t=jpeg")},
			preprocessingParameters{_type: "jpeg", quality: 100, width: 0, height: 0},
			false,
		}, {
			"type png should succeed",
			args{parseQuery("t=png")},
			preprocessingParameters{_type: "png", quality: 100, width: 0, height: 0},
			false,
		}, {
			"t=gif should succeed",
			args{parseQuery("t=gif")},
			preprocessingParameters{_type: "gif", quality: 100, width: 0, height: 0},
			false,
		}, {
			"width and height set",
			args{parseQuery("w=50&h=500")},
			preprocessingParameters{quality: 100, width: 50, height: 500, _type: "jpeg"},
			false,
		}, {
			"type jpg should fail",
			args{parseQuery("w=50&h=500&t=jpg")},
			preprocessingParameters{},
			true,
		}, {
			"type vim should fail",
			args{parseQuery("w=50&t=vim&h=500")},
			preprocessingParameters{},
			true,
		}, {
			"q=-1 should fail",
			args{parseQuery("q=-1")},
			preprocessingParameters{},
			true,
		}, {
			"q=abc should fail",
			args{parseQuery("q=abc")},
			preprocessingParameters{},
			true,
		}, {
			"q=101 should fail",
			args{parseQuery("q=101")},
			preprocessingParameters{},
			true,
		}, {
			"width of 0 should fail",
			args{parseQuery("w=0")},
			preprocessingParameters{},
			true,
		}, {
			"height of 0 should fail",
			args{parseQuery("h=0")},
			preprocessingParameters{},
			true,
		}, {
			"height fails upon recieving string",
			args{parseQuery("h=full")},
			preprocessingParameters{},
			true,
		}, {
			"width fails upon recieving string",
			args{parseQuery("w=full")},
			preprocessingParameters{},
			true,
		}, {
			"Values given are very large",
			args{parseQuery("w=999999999999999999&q=100&h=990000")},
			preprocessingParameters{quality: 100, width: 999999999999999999, height: 990000, _type: "jpeg"},
			false,
		}, {
			"Values given are TOO large",
			args{parseQuery("w=99999999999999999999&q=100&h=990000")},
			preprocessingParameters{},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseParameters(tt.args.v)
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
