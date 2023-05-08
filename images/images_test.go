package images

import (
	"os"
	"testing"
	"time"

	"github.com/charmbracelet/log"
)

func TestSize_String(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want string
	}{
		{"0", Size(0), "0 B"},
		{"7 B", Size(7), "7 B"},
		{"7.78 KB", Size(7*Kilobyte + 777), "7.78 KB"},
		{"7.00 MB", Size(7*Megabyte + 77), "7.00 MB"},
		{"7.75 MB round up", Size(7*Megabyte + 745*Kilobyte), "7.75 MB"},
		{"7.75 MB round down", Size(7*Megabyte + 754*Kilobyte), "7.75 MB"},
		{"900 GB", Size(900 * Gigabyte), "900 GB"},
		{"900.00 GB", Size(900*Gigabyte + 1), "900.00 GB"},
		{"12 TB", Size(12*Terabyte + 1*Gigabyte + 777*Megabyte + 7*Kilobyte + 42), "12.00 TB"},
		{"12.70 PB", Size(12*Petabyte + 695*Terabyte), "12.70 PB"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("Size.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormat_String(t *testing.T) {
	tests := []struct {
		name string
		f    Format
		want string
	}{
		{"Jpeg", Jpeg, "jpeg"},
		{"png", Png, "png"},
		{"gif", Gif, "gif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.String(); got != tt.want {
				t.Errorf("Format.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageHandler_originalPath(t *testing.T) {
	// Config
	conf := Config{OriginalsDir: "originals"}
	type fields struct {
		conf     Config
		latestId ImageId
		l        *log.Logger
		cache    cache
	}
	type args struct {
		id ImageId
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"originals/1.jpg", fields{conf: conf}, args{ImageId(1)}, "originals/1.jpg"},
		{"originals/1.jpg", fields{conf: conf}, args{ImageId(10)}, "originals/10.jpg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ImageHandler{
				conf:     tt.fields.conf,
				latestId: tt.fields.latestId,
				l:        tt.fields.l,
				cache:    tt.fields.cache,
			}
			if got := h.originalPath(tt.args.id); got != tt.want {
				t.Errorf("ImageHandler.originalPath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageHandler_cachePath(t *testing.T) {
	conf := Config{OriginalsDir: "originals"}

	type fields struct {
		conf     Config
		latestId ImageId
		l        *log.Logger
		cache    cache
	}
	type args struct {
		params ImageParameters
		id     ImageId
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{"id 1, jpeg, 100x100", fields{conf: conf}, args{ImageParameters{Format: Jpeg, Width: 100, Height: 100}, ImageId(1)}, "1_100x100_q0_0.jpeg"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &ImageHandler{
				conf:     tt.fields.conf,
				latestId: tt.fields.latestId,
				l:        tt.fields.l,
				cache:    tt.fields.cache,
			}
			if got := h.cachePath(tt.args.params, tt.args.id); got != tt.want {
				t.Errorf("ImageHandler.cachePath() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImageParameters_String(t *testing.T) {
	type fields struct {
		Format  Format
		Width   uint
		Height  uint
		Quality int
		MaxSize Size
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{"jpeg 100x100", fields{Format: Jpeg, Width: 100, Height: 100}, "100x100_q0_0.jpeg"},
		{"gif q256", fields{Format: Gif, Quality: 256}, "0x0_q256_0.gif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageParameters{
				Format:  tt.fields.Format,
				Width:   tt.fields.Width,
				Height:  tt.fields.Height,
				Quality: tt.fields.Quality,
				MaxSize: tt.fields.MaxSize,
			}
			if got := ip.String(); got != tt.want {
				t.Errorf("ImageParameters.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

const (
	testFsDir          = "test-fs"
	test_import_source = testFsDir + "/originals"
)

func Test_CacheHousekeeping(t *testing.T) {

	// Arange
	tempDir, err := os.MkdirTemp(testFsDir, "Test_CacheHousekeeping-")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	h, err := New(Config{
		OriginalsDir: tempDir,
		CacheDir:     tempDir,
		CreateDirs:   true,
		SetPerms:     true,
		CacheRules: CacheRules{
			MaxTotalCacheSize: Size(100 * Kilobyte),
			MaxTimeSinceUse:   60 * time.Second},
	},
		nil,
	)
	if err != nil {
		t.Errorf("New() error = %v", err)
	}

	h.Add(test_import_source + "/one.jpg")
	h.Add(test_import_source + "/two.jpg")
	h.Add(test_import_source + "/three.jpg")
	h.Add(test_import_source + "/four.jpg")

	bytesFreed, err := h.CacheHouseKeeping()
	if err != nil {
		t.Errorf("CacheHouseKeeping() error = %v", err)
	}
	if bytesFreed != 15 {
		t.Errorf("CacheHouseKeeping() size = %v, want %v", bytesFreed, 15)
	}
	// t.FailNow()
}
