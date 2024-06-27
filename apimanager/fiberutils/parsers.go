package fiberutils

import (
	"strconv"
	"time"
)

// Parser is a function that parses a string into a value of the given type.
type Parser[T any] func(string) (T, error)

// IntParser is a type constraint for integer parsers.
type IntParser interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64
}

// ParseInt returns a parser that parses an integer string into the given type.
func ParseInt[T IntParser]() Parser[T] {
	return func(s string) (T, error) {
		v, err := strconv.ParseInt(s, 10, getBitSize(T(0)))
		return T(v), err
	}
}

// UintParser is a type constraint for unsigned integer parsers.
type UintParser interface {
	~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// ParseUint returns a parser that parses an unsigned integer string into the given type.
func ParseUint[T UintParser]() Parser[T] {
	return func(s string) (T, error) {
		v, err := strconv.ParseUint(s, 10, getBitSize(T(0)))
		return T(v), err
	}
}

// FloatParser is a type constraint for float parsers.
type FloatParser interface {
	~float32 | ~float64
}

// ParseFloat returns a parser that parses a float string into the given type.
func ParseFloat[T FloatParser]() Parser[T] {
	return func(s string) (T, error) {
		v, err := strconv.ParseFloat(s, getBitSize(T(0)))
		return T(v), err
	}
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
	bitSize8  = 8
	bitSize16 = 16
	bitSize32 = 32
	bitSize64 = 64
)

// getBitSize returns the bit size for various types.
// If the type is not supported, it returns 0.
func getBitSize(zero any) int {
	switch zero.(type) {
	case int, int64, uint, uint64:
		return bitSize64
	case int8, uint8:
		return bitSize8
	case int16, uint16:
		return bitSize16
	case int32, uint32, float32:
		return bitSize32
	case float64:
		return bitSize64
	default:
		return 0
	}
}
