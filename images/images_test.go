package images

import (
	"testing"
)

func TestSize_String(t *testing.T) {
	tests := []struct {
		name string
		s    Size
		want string
	}{
		{"0", Size(0), "0"},
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
