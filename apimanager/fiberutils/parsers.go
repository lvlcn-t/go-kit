package fiberutils

import (
	"fmt"
	"net"
	"runtime"
	"strconv"
	"time"

	"golang.org/x/exp/constraints"
)

// Parser is a function that parses a string into a value of the given type.
type Parser[T any] func(string) (T, error)

// ParseInt parses an integer string into the given signed type.
func ParseInt[T constraints.Signed](s string) (T, error) {
	v, err := strconv.ParseInt(s, 10, getBitSize(T(0)))
	return T(v), err
}

// ParseUint parses an unsigned integer string into the given unsigned type.
func ParseUint[T constraints.Unsigned](s string) (T, error) {
	v, err := strconv.ParseUint(s, 10, getBitSize(T(0)))
	return T(v), err
}

// ParseFloat parses a float string into the given float type.
func ParseFloat[T constraints.Float](s string) (T, error) {
	v, err := strconv.ParseFloat(s, getBitSize(T(0)))
	return T(v), err
}

// ParseComplex parses a complex string into the given complex type.
func ParseComplex[T constraints.Complex](s string) (T, error) {
	v, err := strconv.ParseComplex(s, getBitSize(T(0)))
	return T(v), err
}

// ParseIP parses an IP address string into a [net.IP].
func ParseIP(s string) (net.IP, error) {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil, net.InvalidAddrError(fmt.Sprintf("invalid address: %q", s))
	}
	return ip, nil
}

// ParseDate returns a parser that parses a date string into a time.Time using the given formats.
// The first format that successfully parses the date will be used.
// If no formats are provided, the default format [time.DateOnly] will be used.
func ParseDate(format ...string) Parser[time.Time] {
	if len(format) == 0 {
		format = append(format, time.DateOnly)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}

// ParseTime returns a parser that parses a time string into a time.Time using the given formats.
// The first format that successfully parses the time will be used.
// If no formats are provided, the default format [time.TimeOnly] will be used.
func ParseTime(format ...string) Parser[time.Time] {
	if len(format) == 0 {
		format = append(format, time.TimeOnly)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}

// ParseDateTime returns a parser that parses a date and time string into a time.Time using the given formats.
// The first format that successfully parses the date and time will be used.
// If no formats are provided, the default format [time.RFC3339] will be used.
func ParseDateTime(format ...string) Parser[time.Time] {
	if len(format) == 0 {
		format = append(format, time.RFC3339)
	}

	return func(s string) (t time.Time, err error) {
		for _, f := range format {
			t, err = time.Parse(f, s)
			if err == nil {
				return t, nil
			}
		}
		return t, err
	}
}

// bitSize constants for various types.
const (
	bitSize8 int = 1 << (iota + 3)
	bitSize16
	bitSize32
	bitSize64
	bitSize128
)

// getBitSize returns the bit size for various types.
// If the type is not supported, it returns 0.
func getBitSize(zero any) int {
	switch zero.(type) {
	case int8, uint8:
		return bitSize8
	case int16, uint16:
		return bitSize16
	case int32, uint32, float32:
		return bitSize32
	case int64, uint64, float64, complex64:
		return bitSize64
	case int, uint, uintptr:
		if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
			return bitSize32
		}
		return bitSize64
	case complex128:
		return bitSize128
	default:
		return 0
	}
}
