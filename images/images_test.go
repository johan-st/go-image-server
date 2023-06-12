package images

import (
	"testing"
)

func TestSize_String(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		s    Size
		want string
	}{
		{"0", Size(0), "0 B"},
		{"7 B", Size(7), "7 B"},
		{"7.76 KB", Size(7*Kilobyte + 777), "7.76 KB"},
		{"7.00 MB", Size(7*Megabyte + 77), "7.00 MB"},
		{"7.73 MB round up", Size(7*Megabyte + 752*Kilobyte), "7.73 MB"},
		{"7.74 MB round down", Size(7*Megabyte + 754*Kilobyte), "7.74 MB"},
		{"900 GB", Size(900 * Gigabyte), "900 GB"},
		{"900.00 GB", Size(900*Gigabyte + 1), "900.00 GB"},
		{"12.00 TB", Size(12*Terabyte + 1*Gigabyte + 777*Megabyte + 7*Kilobyte + 42), "12.00 TB"},
		{"12.68 PB", Size(12*Petabyte + 695*Terabyte), "12.68 PB"},

		// Observed issues
		{"3 decimal bug", Size(4*Megabyte + 1023*Kilobyte), "5.00 MB"},
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

func TestImageParameters_String(t *testing.T) {
	t.Parallel()
	type fields struct {
		Id      int
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
		{"jpeg 100x100", fields{Id: 42, Format: Jpeg, Width: 100, Height: 100}, "42_100x100_q0_s0.jpeg"},
		{"gif q256", fields{Id: 9, Format: Gif, Quality: 256}, "9_0x0_q256_s0.gif"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := &ImageParameters{
				Id:      tt.fields.Id,
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

// const (
// 	testFsDir          = "test-fs"
// 	test_import_source = testFsDir + "/originals"
// )

// func Test_CacheHousekeeping(t *testing.T) {

// 	// Arange
// 	tempDir, err := os.MkdirTemp(testFsDir, "Test_CacheHousekeeping-")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	defer os.RemoveAll(tempDir)

// 	h, err := New(Config{
// 		OriginalsDir: tempDir,
// 		CacheDir:     tempDir,
// 		CreateDirs:   true,
// 		SetPerms:     true,
// 	},
// 		nil,
// 	)
// 	if err != nil {
// 		t.Errorf("New() error = %v", err)
// 	}

// 	h.Add(test_import_source + "/one.jpg")
// 	h.Add(test_import_source + "/two.jpg")
// 	h.Add(test_import_source + "/three.jpg")
// 	h.Add(test_import_source + "/four.jpg")

// 	bytesFreed, err := h.CacheHouseKeeping()
// 	if err != nil {
// 		t.Errorf("CacheHouseKeeping() error = %v", err)
// 	}
// 	if bytesFreed != 15 {
// 		t.Errorf("CacheHouseKeeping() size = %v, want %v", bytesFreed, 15)
// 	}
// 	// t.FailNow()
// }

func Test_SizeFromString(t *testing.T) {
	t.Parallel()

	type args struct {
		str string
	}
	tests := []struct {
		name    string
		args    args
		want    Size
		wantErr bool
	}{
		// Success
		{"0", args{"0"}, Size(0), false},
		{"50", args{"50"}, Size(50), false},
		{"1 B", args{"1 B"}, Size(1), false},
		{"5 KB", args{"5 KB"}, 5 * Kilobyte, false},
		{"10 MB", args{"10 MB"}, 10 * Megabyte, false},
		{"15GB", args{"15GB"}, 15 * Gigabyte, false},
		{"20 TB", args{"20 TB"}, 20 * Terabyte, false},
		{"25 PB", args{"25 PB"}, 25 * Petabyte, false},

		// Failure
		{"-1", args{"-1"}, 0, true},
		{"1.5", args{"1.5"}, 0, true},
		{"1.5 B", args{"1.5 B"}, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSize(tt.args.str)
			if (err != nil) != tt.wantErr {
				t.Errorf("Size.FromString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Size.FromString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_ratioToPixels(t *testing.T) {
	type args struct {
		ratio  float64
		width  float64
		height float64
	}
	tests := []struct {
		name  string
		args  args
		wantW int
		wantH int
	}{
		{"trivial case", args{1, 0, 0}, 0, 0},
		{"return same", args{2, 100, 50}, 100, 50},
		{"return cropped, quadratic", args{1, 100, 41}, 41, 41},
		{"height bound portrait", args{0.5, 100, 40}, 20, 40},
		{"width bound portrait", args{0.5, 50, 120}, 50, 100},
		{"height bound landscape", args{1.5, 100, 60}, 90, 60},
		{"width bound landscape", args{3, 300, 150}, 300, 100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotW, gotH := ratioToPixels(tt.args.ratio, tt.args.width, tt.args.height)
			if gotW != tt.wantW {
				t.Errorf("ratioToPixels() gotW = %v, want %v", gotW, tt.wantW)
			}
			if gotH != tt.wantH {
				t.Errorf("ratioToPixels() gotH = %v, want %v", gotH, tt.wantH)
			}
		})
	}
}
