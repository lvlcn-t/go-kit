package lists

import (
	"reflect"
)

// Predicate is a function that returns true if the element should be included in the new slice.
type Predicate[T any] func(T) bool

// Filter returns a new slice containing only the elements of the original slice that satisfy the predicate f.
// If no elements satisfy the predicate, an empty slice is returned.
//
// Example:
//
//	Filter([]int{1, 2, 3, 4}, func(i int) bool {
//		return i%2 == 0
//	}) // Output: []int{2, 4}
func Filter[T any](slice []T, f Predicate[T]) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if f(item) {
			result = append(result, item)
		}
	}
	return result
}

// FilterEmpty returns a new slice containing only the non-zero elements of the original slice.
//
// Example:
//
//	FilterEmpty([]int{1, 2, 3, 0, 4}) // Output: []int{1, 2, 3, 4}
func FilterEmpty[T any](slice []T) []T {
	return Filter(slice, func(item T) bool {
		return !reflect.ValueOf(item).IsZero()
	})
}

// FilterNil returns a new slice containing only the non-nil elements of the original slice.
//
// Example:
//
//	FilterNil([]*int{nil, new(int), nil}) // Output: []*int{new(int)}
func FilterNil[T any](slice []T) []T {
	return Filter(slice, func(item T) bool {
		return !reflect.ValueOf(item).IsNil()
	})
}

// FilterNonEmpty returns a new slice containing only the zero elements of the original slice.
//
// Example:
//
//	FilterNonEmpty([]int{1, 2, 3, 0, 4}) // Output: []int{0}
func FilterNonEmpty[T any](slice []T) []T {
	return Filter(slice, func(item T) bool {
		return reflect.ValueOf(item).IsZero()
	})
}

// FilterNonNil returns a new slice containing only the nil elements of the original slice.
//
// Example:
//
//	FilterNonNil([]*int{nil, new(int), nil}) // Output: []*int{nil, nil}
func FilterNonNil[T any](slice []T) []T {
	return Filter(slice, func(item T) bool {
		return reflect.ValueOf(item).IsNil()
	})
}
