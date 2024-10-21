package env

import (
	"fmt"
	"reflect"
	"runtime"
	"strconv"
)

// Converter is a function that converts a string to a desired type.
type Converter[T any] func(string) (T, error)

// defaultConverter returns a default converter for the provided type.
func defaultConverter[T any]() (Converter[T], error) {
	typ := reflect.TypeFor[T]()

	switch typ.Kind() {
	case reflect.String:
		return func(v string) (T, error) {
			return any(v).(T), nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertInt[T], nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return convertUint[T], nil
	case reflect.Float32, reflect.Float64:
		return convertFloat[T], nil
	case reflect.Complex64, reflect.Complex128:
		return convertComplex[T], nil
	case reflect.Bool:
		return func(s string) (T, error) {
			v, err := strconv.ParseBool(s)
			return any(v).(T), err
		}, nil
	default:
		return nil, fmt.Errorf("type %v is not supported", typ)
	}
}

// convertInt converts a string to an integer.
func convertInt[T any](v string) (T, error) {
	var zero T
	i, err := strconv.ParseInt(v, 10, getBitSize(zero))
	if err != nil {
		return zero, err
	}
	return reflect.ValueOf(i).Convert(reflect.TypeOf(zero)).Interface().(T), nil
}

// convertUint converts a string to an unsigned integer.
func convertUint[T any](v string) (T, error) {
	var zero T
	i, err := strconv.ParseUint(v, 10, getBitSize(zero))
	if err != nil {
		return zero, err
	}
	return reflect.ValueOf(i).Convert(reflect.TypeOf(zero)).Interface().(T), nil
}

// convertFloat converts a string to a float.
func convertFloat[T any](v string) (T, error) {
	var zero T
	f, err := strconv.ParseFloat(v, getBitSize(zero))
	if err != nil {
		return zero, err
	}
	return reflect.ValueOf(f).Convert(reflect.TypeOf(zero)).Interface().(T), nil
}

// convertComplex converts a string to a complex number.
func convertComplex[T any](v string) (T, error) {
	var zero T
	c, err := strconv.ParseComplex(v, getBitSize(zero))
	if err != nil {
		return zero, err
	}
	return reflect.ValueOf(c).Convert(reflect.TypeOf(zero)).Interface().(T), nil
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
