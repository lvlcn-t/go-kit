package lists

import (
	"math/rand"
	"time"
)

// Map applies the function f to each element of the slice, returning a new slice with the results.
//
// Example:
//
//	Map([]int{1, 2, 3}, func(i int) int {
//		return i * i
//	}) // Output: []int{1, 4, 9}
func Map[T any, R any](slice []T, f func(T) R) []R {
	var result []R
	for _, item := range slice {
		result = append(result, f(item))
	}
	return result
}

// Reduce applies the function f to each element of the slice, accumulating the results.
// The accumulated value is initialized with the initial value.
//
// Example:
//
//	Reduce([]int{1, 2, 3}, 0, func(acc, item int) int {
//		return acc + item
//	}) // Output: 6
func Reduce[T any, R any](slice []T, initial R, f func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = f(result, item)
	}
	return result
}

// LastIndexOf returns the index of the last occurrence of the value in the slice, or -1 if the value is not found.
//
// Example:
//
//	LastIndexOf([]int{1, 2, 3, 2}, 2) // Output: 3
func LastIndexOf[T comparable](slice []T, value T) int {
	for i := len(slice) - 1; i >= 0; i-- {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

// Count returns the number of elements in the slice that satisfy the predicate f.
//
// Example:
//
//	Count([]int{1, 2, 3, 4}, func(i int) bool {
//		return i%2 == 0
//	}) // Output: 2
func Count[T any](slice []T, f Predicate[T]) int {
	count := 0
	for _, item := range slice {
		if f(item) {
			count++
		}
	}
	return count
}

// Distinct returns a new slice containing only the unique elements of the original slice.
//
// Example:
//
//	Distinct([]int{1, 2, 3, 2, 1}) // Output: []int{1, 2, 3}
func Distinct[T comparable](slice []T) []T {
	seen := make(map[T]struct{})
	var result []T
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
//
// Example:
//
//	Partition([]int{1, 2, 3, 4}, func(i int) bool {
//		return i%2 == 0
//	}) // Output: []int{2, 4}, []int{1, 3}
func Partition[T any](slice []T, f Predicate[T]) ([]T, []T) {
	var trueSlice, falseSlice []T
	for _, v := range slice {
		if f(v) {
			trueSlice = append(trueSlice, v)
		} else {
			falseSlice = append(falseSlice, v)
		}
	}
	return trueSlice, falseSlice
}

// Permutations returns all possible permutations of the slice.
//
// Example:
//
//	Permutations([]int{1, 2, 3}) // Output: [][]int{{1, 2, 3}, {2, 1, 3}, {3, 1, 2}, {1, 3, 2}, {2, 3, 1}, {3, 2, 1}}
func Permutations[T any](slice []T) [][]T {
	var arrangeElements func([]T, int)
	result := [][]T{}
	arrangeElements = func(arr []T, n int) {
		if n == 1 {
			tmp := make([]T, len(arr))
			copy(tmp, arr)
			result = append(result, tmp)
			return
		}

		for i := 0; i < n; i++ {
			arrangeElements(arr, n-1)
			if n%2 == 1 {
				arr[i], arr[n-1] = arr[n-1], arr[i]
				continue
			}
			arr[0], arr[n-1] = arr[n-1], arr[0]
		}
	}
	arrangeElements(slice, len(slice))
	return result
}

// Combinations returns all possible combinations of n elements from the slice.
//
// Example:
//
//	Combinations([]int{1, 2, 3}, 2) // Output: [][]int{{1, 2}, {1, 3}, {2, 3}}
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
//
// Example:
//
//	Shuffle([]int{1, 2, 3, 4}) // Output: []int{3, 1, 4, 2}
func Shuffle[T any](slice []T) []T {
	result := make([]T, len(slice))
	perm := rand.New(rand.NewSource(time.Now().UnixNano())).Perm(len(slice))
	for i, v := range perm {
		result[v] = slice[i]
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
//
// Example:
//
//	Zip([]int{1, 2, 3}, []string{"a", "b", "c"}) // Output: []Pair{{1, "a"}, {2, "b"}, {3, "c"}}
func Zip[T any, R any](slice1 []T, slice2 []R) []Pair[T, R] {
	var result []Pair[T, R]
	for i := 0; i < len(slice1) && i < len(slice2); i++ {
		result = append(result, Pair[T, R]{First: slice1[i], Second: slice2[i]})
	}
	return result
}

// Unzip converts a slice of pairs into a pair of slices.
//
// Example:
//
//	Unzip([]Pair{{1, "a"}, {2, "b"}, {3, "c"}}) // Output: []int{1, 2, 3}, []string{"a", "b", "c"}
func Unzip[T any, R any](pairs []Pair[T, R]) ([]T, []R) {
	var slice1 []T
	var slice2 []R
	for _, p := range pairs {
		slice1 = append(slice1, p.First)
		slice2 = append(slice2, p.Second)
	}
	return slice1, slice2
}

// Chunk splits the slice into chunks of the specified size.
//
// Example:
//
//	Chunk([]int{1, 2, 3, 4, 5}, 2) // Output: [][]int{{1, 2}, {3, 4}, {5}}
func Chunk[T any](slice []T, size int) [][]T {
	var result [][]T
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
//
// Example:
//
//	Flatten([][]int{{1, 2}, {3, 4}, {5}}) // Output: []int{1, 2, 3, 4, 5}
func Flatten[T any](slices [][]T) []T {
	var result []T
	for _, slice := range slices {
		result = append(result, slice...)
	}
	return result
}

// Intersect returns a new slice containing only the elements that are present in all input slices.
//
// Example:
//
//	Intersect([]int{1, 2, 3}, []int{2, 3, 4}, []int{3, 4, 5}) // Output: []int{3}
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

	var result []T
	for k, v := range counts {
		if v == len(slices) {
			result = append(result, k)
		}
	}
	return result
}

// Difference returns a new slice containing the elements that are present in the first slice but not in the other input slices.
//
// Example:
//
//	Difference([]int{1, 2, 3}, []int{2, 3, 4}, []int{3, 4, 5}) // Output: []int{1}
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

	var result []T
	for _, v := range slices[0] {
		if _, ok := seen[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}

// Union returns a new slice containing all the unique elements that are present in any of the input slices.
//
// Example:
//
//	Union([]int{1, 2, 3}, []int{2, 3, 4}, []int{3, 4, 5}) // Output: []int{1, 2, 3, 4, 5}
func Union[T comparable](slices ...[]T) []T {
	seen := make(map[T]struct{})
	var result []T
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
//
// Example:
//
//	IsSorted([]int{1, 2, 3}, func(a, b int) bool {
//		return a < b
//	}) // Output: true
func IsSorted[T comparable](slice []T, f func(T, T) bool) bool {
	for i := 1; i < len(slice); i++ {
		if f(slice[i], slice[i-1]) {
			return false
		}
	}
	return true
}

// AllMatch returns true if all elements of the slice satisfy the predicate f.
//
// Example:
//
//	AllMatch([]int{1, 2, 3, 4}, func(i int) bool {
//		return i > 0
//	}) // Output: true
func AllMatch[T any](slice []T, f Predicate[T]) bool {
	for _, item := range slice {
		if !f(item) {
			return false
		}
	}
	return true
}

// AnyMatch returns true if any element of the slice satisfies the predicate f.
//
// Example:
//
//	AnyMatch([]int{1, 2, 3, 4}, func(i int) bool {
//		return i%2 == 0
//	}) // Output: true
func AnyMatch[T any](slice []T, f Predicate[T]) bool {
	for _, item := range slice {
		if f(item) {
			return true
		}
	}
	return false
}

// NoneMatch returns true if none of the elements of the slice satisfy the predicate f.
//
// Example:
//
//	NoneMatch([]int{1, 2, 3, 4}, func(i int) bool {
//		return i < 0
//	}) // Output: true
func NoneMatch[T any](slice []T, f Predicate[T]) bool {
	return !AnyMatch(slice, f)
}
