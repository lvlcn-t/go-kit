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
	pointer := typ.Kind() == reflect.Pointer
	if pointer {
		typ = typ.Elem()
	}

	switch typ.Kind() {
	case reflect.String:
		return func(v string) (T, error) {
			if pointer {
				return any(&v).(T), nil
			}
			return any(v).(T), nil
		}, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertInt[T](pointer), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return convertUint[T](pointer), nil
	case reflect.Float32, reflect.Float64:
		return convertFloat[T](pointer), nil
	case reflect.Complex64, reflect.Complex128:
		return convertComplex[T](pointer), nil
	case reflect.Bool:
		return func(s string) (T, error) {
			v, err := strconv.ParseBool(s)
			if pointer {
				return any(&v).(T), err
			}
			return any(v).(T), err
		}, nil
	default:
		return nil, fmt.Errorf("type %v is not supported", typ)
	}
}

// convertInt converts a string to an integer.
func convertInt[T any](pointer bool) Converter[T] {
	return func(v string) (T, error) {
		var zero T
		i, err := strconv.ParseInt(v, 10, getBitSize(zero))
		if err != nil {
			return zero, err
		}
		if pointer {
			ptrValue := newPointerForType(zero)
			ptrValue.Elem().SetInt(i)
			return ptrValue.Interface().(T), nil
		}
		return reflect.ValueOf(i).Convert(reflect.TypeOf(zero)).Interface().(T), nil
	}
}

// newPointerForType creates a new pointer for the provided type.
func newPointerForType[T any](zero T) reflect.Value {
	return reflect.New(reflect.TypeOf(zero).Elem())
}

// convertUint converts a string to an unsigned integer.
func convertUint[T any](pointer bool) Converter[T] {
	return func(v string) (T, error) {
		var zero T
		i, err := strconv.ParseUint(v, 10, getBitSize(zero))
		if err != nil {
			return zero, err
		}
		if pointer {
			ptrValue := newPointerForType(zero)
			ptrValue.Elem().SetUint(i)
			return ptrValue.Interface().(T), nil
		}
		return reflect.ValueOf(i).Convert(reflect.TypeOf(zero)).Interface().(T), nil
	}
}

// convertFloat converts a string to a float.
func convertFloat[T any](pointer bool) Converter[T] {
	return func(v string) (T, error) {
		var zero T
		f, err := strconv.ParseFloat(v, getBitSize(zero))
		if err != nil {
			return zero, err
		}
		if pointer {
			ptrValue := newPointerForType(zero)
			ptrValue.Elem().SetFloat(f)
			return ptrValue.Interface().(T), nil
		}
		return reflect.ValueOf(f).Convert(reflect.TypeOf(zero)).Interface().(T), nil
	}
}

// convertComplex converts a string to a complex number.
func convertComplex[T any](pointer bool) Converter[T] {
	return func(v string) (T, error) {
		var zero T
		c, err := strconv.ParseComplex(v, getBitSize(zero))
		if err != nil {
			return zero, err
		}
		if pointer {
			ptrValue := newPointerForType(zero)
			ptrValue.Elem().SetComplex(c)
			return ptrValue.Interface().(T), nil
		}
		return reflect.ValueOf(c).Convert(reflect.TypeOf(zero)).Interface().(T), nil
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
	case int8, uint8, *int8, *uint8:
		return bitSize8
	case int16, uint16, *int16, *uint16:
		return bitSize16
	case int32, uint32, float32, *int32, *uint32, *float32:
		return bitSize32
	case int64, uint64, float64, complex64, *int64, *uint64, *float64, *complex64:
		return bitSize64
	case int, uint, uintptr, *int, *uint, *uintptr:
		if runtime.GOARCH == "386" || runtime.GOARCH == "arm" {
			return bitSize32
		}
		return bitSize64
	case complex128, *complex128:
		return bitSize128
	default:
		return 0
	}
}
