package lists

import (
	"fmt"
	"math/rand/v2"
)

// Apply applies the function f to each element of the slice, returning a new slice with the results.
func Apply[T any, R any](slice []T, f func(T) R) []R {
	result := make([]R, 0, len(slice))
	for _, item := range slice {
		result = append(result, f(item))
	}
	return result
}

// Reduce applies the function f to each element of the slice, accumulating the results.
// The accumulated value is initialized with the initial value.
func Reduce[T any, R any](slice []T, initial R, f func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = f(result, item)
	}
	return result
}

// LastIndexOf returns the index of the last occurrence of the value in the slice, or -1 if the value is not found.
func LastIndexOf[T comparable](slice []T, value T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

// Count returns the number of elements in the slice that satisfy the predicate f.
func Count[T any](slice []T, f Predicate[T]) int {
	count := 0
	for _, item := range slice {
		if f(item) {
			count++
		}
	}
	return count
}

// Counter represents a map with the count of each element.
type Counter[T comparable] map[T]int

func (c Counter[T]) String() string {
	return fmt.Sprintf("%v", map[T]int(c))
}

// CountBy returns a map with the count of each element in the slice.
func CountBy[T comparable](slice []T) Counter[T] {
	counts := make(Counter[T], len(slice))
	for _, item := range slice {
		counts[item]++
	}
	return counts
}

// Get returns the count of the element in the counter.
func (c Counter[T]) Get(key T) int {
	return c[key]
}

// MostCommon returns the most common element(s) in the counter.
func (c Counter[T]) MostCommon() []T {
	result := []T{}
	maximum := 0
	for k, v := range c {
		if v > maximum {
			result = []T{k}
			maximum = v
			continue
		}
		if v == maximum {
			result = append(result, k)
		}
	}
	return result
}

// LeastCommon returns the least common element(s) in the counter.
func (c Counter[T]) LeastCommon() []T {
	result := []T{}
	minimum := 0
	for k, v := range c {
		if minimum == 0 || v < minimum {
			result = []T{k}
			minimum = v
			continue
		}
		if v == minimum {
			result = append(result, k)
		}
	}
	return result
}

// Elements returns a slice with the elements in the counter.
func (c Counter[T]) Elements() []T {
	result := []T{}
	for k := range c {
		result = append(result, k)
	}
	return result
}

// Total returns the total count of all elements in the counter.
func (c Counter[T]) Total() int {
	total := 0
	for _, v := range c {
		total += v
	}
	return total
}

// Clear removes all elements from the counter.
func (c Counter[T]) Clear() {
	for k := range c {
		delete(c, k)
	}
}

// Distinct returns a new slice containing only the unique elements of the original slice.
func Distinct[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0, len(slice))
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// Partition returns two slices, the first containing the elements of the original slice that satisfy the predicate f,
// and the second containing the elements that do not satisfy the predicate.
func Partition[T any](slice []T, f Predicate[T]) (matched, unmatched []T) {
	matched = make([]T, 0, len(slice))
	unmatched = make([]T, 0, len(slice))
	for _, v := range slice {
		if f(v) {
			matched = append(matched, v)
			continue
		}
		unmatched = append(unmatched, v)
	}
	return matched, unmatched
}

// Permutations returns all possible permutations of the slice.
func Permutations[T any](slice []T) [][]T {
	var permute func([]T, int)
	result := [][]T{}
	permute = func(arr []T, n int) {
		if n == 1 {
			tmp := make([]T, len(arr))
			copy(tmp, arr)
			result = append(result, tmp)
			return
		}

		for i := 0; i < n; i++ {
			permute(arr, n-1)
			if n%2 == 1 {
				arr[i], arr[n-1] = arr[n-1], arr[i]
				continue
			}
			arr[0], arr[n-1] = arr[n-1], arr[0]
		}
	}
	permute(slice, len(slice))
	return result
}

// Combinations returns all possible combinations of n elements from the slice.
func Combinations[T any](slice []T, n int) [][]T {
	var result [][]T
	var combine func([]T, int, int, []T)
	combine = func(arr []T, start int, k int, data []T) {
		if k == 0 {
			tmp := make([]T, len(data))
			copy(tmp, data)
			result = append(result, tmp)
			return
		}

		for i := start; i <= len(arr)-k; i++ {
			data[len(data)-k] = arr[i]
			combine(arr, i+1, k-1, data)
		}
	}
	data := make([]T, n)
	combine(slice, 0, n, data)
	return result
}

// Shuffle returns a new slice with the elements of the original slice shuffled.
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	perm := rand.Perm(len(slice)) // #nosec G404 // No need for cryptographically secure random number generator
	for i, v := range perm {
		result[v] = slice[i]
	}
	return result
}

// Sample returns a new slice with n elements randomly sampled from the original slice.
func Sample[T any](slice []T, n int) []T {
	if n >= len(slice) {
		return slice
	}

	result := make([]T, n)
	perm := rand.Perm(len(slice)) // #nosec G404 // No need for cryptographically secure random number generator
	for i := 0; i < n; i++ {
		result[i] = slice[perm[i]]
	}
	return result
}

// Pair represents a pair of elements.
type Pair[T any, R any] struct {
	First  T
	Second R
}

// Zip combines the elements from two slices into a single slice of pairs.
// The resulting slice has the length of the shortest input slice.
func Zip[T any, R any](slice1 []T, slice2 []R) []Pair[T, R] {
	l := len(slice1)
	if len(slice2) < l {
		l = len(slice2)
	}
	result := make([]Pair[T, R], 0, l)
	for i := 0; i < len(slice1) && i < len(slice2); i++ {
		result = append(result, Pair[T, R]{First: slice1[i], Second: slice2[i]})
	}
	return result
}

// Unzip converts a slice of pairs into a pair of slices.
func Unzip[T any, R any](pairs []Pair[T, R]) (p1 []T, p2 []R) {
	slice1 := make([]T, 0, len(pairs))
	slice2 := make([]R, 0, len(pairs))
	for _, p := range pairs {
		slice1 = append(slice1, p.First)
		slice2 = append(slice2, p.Second)
	}
	return slice1, slice2
}

// Chunk splits the slice into chunks of the specified size.
func Chunk[T any](slice []T, size int) [][]T {
	result := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		result = append(result, slice[i:end])
	}
	return result
}

// Flatten flattens the nested slices into a single slice.
func Flatten[T any](slices [][]T) []T {
	result := make([]T, 0)
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

// Intersect returns a new slice containing only the elements that are present in all input slices.
func Intersect[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return slices[0]
	}

	counts := make(map[T]int)
	for _, slice := range slices {
		seen := make(map[T]struct{})
		for _, v := range slice {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				counts[v]++
			}
		}
	}

	result := make([]T, 0)
	for k, v := range counts {
		if v == len(slices) {
			result = append(result, k)
		}
	}
	return result
}

// Difference returns a new slice containing the elements that are present in the first slice but not in the other input slices.
func Difference[T comparable](slices ...[]T) []T {
	if len(slices) == 0 {
		return []T{}
	}
	if len(slices) == 1 {
		return slices[0]
	}

	seen := make(map[T]struct{})
	for _, v := range slices[1] {
		seen[v] = struct{}{}
	}

	result := make([]T, 0)
	for _, v := range slices[0] {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// Union returns a new slice containing all the unique elements that are present in any of the input slices.
func Union[T comparable](slices ...[]T) []T {
	seen := make(map[T]struct{})
	result := make([]T, 0)
	for _, slice := range slices {
		for _, v := range slice {
			if _, ok := seen[v]; !ok {
				seen[v] = struct{}{}
				result = append(result, v)
			}
		}
	}
	return result
}

// IsSorted checks if the slice is sorted to the order specified by the comparison function f.
func IsSorted[T comparable](slice []T, f func(T, T) bool) bool {
	for i := 1; i < len(slice); i++ {
		if f(slice[i], slice[i-1]) {
			return false
		}
	}
	return true
}

// AllMatch returns true if all elements of the slice satisfy the predicate f.
func AllMatch[T any](slice []T, f Predicate[T]) bool {
	for _, item := range slice {
		if !f(item) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any element of the slice satisfies the predicate f.
func AnyMatch[T any](slice []T, f Predicate[T]) bool {
	for _, item := range slice {
		if f(item) {
			return true
		}
	}
	return false
}

// NoneMatch returns true if none of the elements of the slice satisfy the predicate f.
func NoneMatch[T any](slice []T, f Predicate[T]) bool {
	return !AnyMatch(slice, f)
}
