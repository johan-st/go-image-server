package main

import "testing"

func Test_pathById(t *testing.T) {
	type args struct {
		id int
	}
	tests := []struct {
		name string
		args args
		want string
	}{{"one", args{1}, "originals/1.jpg"}}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := pathById(tt.args.id); got != tt.want {
				t.Errorf("pathById() = %v, want %v", got, tt.want)
			}
		})
	}
}
