package size

import (
	"fmt"
	"strconv"
	"strings"
)

// Represents file sizes:
//   - 1 Kilobyte = 1024 bytes
//   - 1 Megabyte = 1024 Kilobytes
//   - 1 Gigabyte = 1024 Megabytes
//   - 1 Terabyte = 1024 Gigabytes
//   - 1 Petabyte = 1024 Terabytes
type S uint64

const (
	Kilobyte = 1024            // 1 Kilobyte = 1024 bytes
	Megabyte = 1024 * Kilobyte // 1 Megabyte = 1024 Kilobytes
	Gigabyte = 1024 * Megabyte // 1 Gigabyte = 1024 Megabytes
	Terabyte = 1024 * Gigabyte // 1 Terabyte = 1024 Gigabytes
	Petabyte = 1024 * Terabyte // 1 Petabyte = 1024 Terabytes
)

func Parse(str string) (S, error) {
	if s, ok := sizeParseHelper(str, "KB", Kilobyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "MB", Megabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "GB", Gigabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "TB", Terabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "PB", Petabyte); ok {
		return s, nil
	}
	if s, ok := sizeParseHelper(str, "B", 1); ok {
		return s, nil
	}

	// no suffix, assume bytes
	if intVal, err := strconv.Atoi(str); err == nil && intVal >= 0 {
		return S(intVal), nil
	}

	return S(0), fmt.Errorf("could not parse string '%s' as a Size", str)
}

func sizeParseHelper(str string, sufix string, size S) (S, bool) {
	before, found := strings.CutSuffix(str, sufix)
	if found {
		before = strings.Trim(before, " ")
		intVal, err := strconv.Atoi(before)
		if err != nil || intVal < 0 {
			return S(0), false
		}
		return S(intVal) * size, true
	}
	return S(0), false
}

// TODO: Break out size into a units package
// Size.String()
func (s S) String() string {
	if s == 0 {
		return "0 B"
	}
	if s >= Petabyte {
		return sizeStringHelper(Petabyte, "PB", uint64(s))
	}
	if s >= Terabyte {
		return sizeStringHelper(Terabyte, "TB", uint64(s))
	}
	if s >= Gigabyte {
		return sizeStringHelper(Gigabyte, "GB", uint64(s))
	}
	if s >= Megabyte {
		return sizeStringHelper(Megabyte, "MB", uint64(s))
	}
	if s >= Kilobyte {
		return sizeStringHelper(Kilobyte, "KB", uint64(s))
	}
	return fmt.Sprintf("%d B", s)

}

// sizeStringHelper returnes a string representing the size with two decimals.
// If the amount is exact then it will not show any decimals.
func sizeStringHelper(divInteger int, suffix string, s uint64) string {
	if s%uint64(divInteger) == 0 {
		return fmt.Sprintf("%d %s", s/uint64(divInteger), suffix)
	}
	f := float64(s) / float64(divInteger)
	return fmt.Sprintf("%.2f %s", f, suffix)
}
