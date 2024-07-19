package lists_test

import (
	"fmt"

	"github.com/lvlcn-t/go-kit/lists"
)

func ExampleApply() {
	// Apply the square function to each element
	result := lists.Apply([]int{1, 2, 3}, func(i int) int {
		return i * i
	})

	fmt.Println(result)
	// Output: [1 4 9]
}

func ExampleReduce() {
	// Sum all elements
	result := lists.Reduce([]int{1, 2, 3}, 0, func(acc, item int) int {
		return acc + item
	})

	fmt.Println(result)
	// Output: 6
}

func ExampleLastIndexOf() {
	// Find the last index of 2
	result := lists.LastIndexOf([]int{1, 2, 3, 2}, 2)

	fmt.Println(result)
	// Output: 3
}

func ExampleCount() {
	// Count even numbers
	result := lists.Count([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(result)
	// Output: 2
}

func ExampleDistinct() {
	// Remove duplicates
	result := lists.Distinct([]int{1, 2, 3, 2, 1})

	fmt.Println(result)
	// Output: [1 2 3]
}

func ExamplePartition() {
	// Partition odd and even numbers
	odd, even := lists.Partition([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 != 0
	})

	fmt.Println(odd)
	fmt.Println(even)
	// Output: [1 3]
	// [2 4]
}

func ExamplePermutations() {
	// Generate all permutations
	result := lists.Permutations([]int{1, 2, 3})

	fmt.Println(result)
	// Output: [[1 2 3] [2 1 3] [3 2 1] [2 3 1] [3 1 2] [1 3 2]]
}

func ExampleCombinations() {
	// Generate all combinations
	result := lists.Combinations([]int{1, 2, 3}, 2)

	fmt.Println(result)
	// Output: [[1 2] [1 3] [2 3]]
}

func ExampleShuffle() {
	// Shuffle the slice
	result := lists.Shuffle([]int{1, 2, 3, 4})

	fmt.Println(result)
}

func ExampleZip() {
	// Zip two slices
	result := lists.Zip([]int{1, 2, 3}, []string{"a", "b", "c"})

	fmt.Println(result)
	// Output: [{1 a} {2 b} {3 c}]
}

func ExampleUnzip() {
	zip := []lists.Pair[int, string]{{1, "a"}, {2, "b"}, {3, "c"}}
	// Unzip a slice of pairs into two slices
	result1, result2 := lists.Unzip(zip)

	fmt.Println(result1)
	fmt.Println(result2)
	// Output: [1 2 3]
	// [a b c]
}

func ExampleChunk() {
	// Chunk the slice into smaller slices
	result := lists.Chunk([]int{1, 2, 3, 4, 5}, 2)

	fmt.Println(result)
	// Output: [[1 2] [3 4] [5]]
}

func ExampleFlatten() {
	// Flatten nested slices
	result := lists.Flatten([][]int{{1, 2}, {3, 4}, {5}})

	fmt.Println(result)
	// Output: [1 2 3 4 5]
}

func ExampleIntersect() {
	// Intersect two slices
	result := lists.Intersect([]int{1, 2, 3}, []int{2, 3, 4})

	fmt.Println(result)
	// Output: [2 3]
}

func ExampleDifference() {
	// Find the difference between two slices
	result := lists.Difference([]int{1, 2, 3}, []int{2, 3, 4})

	fmt.Println(result)
	// Output: [1]
}

func ExampleUnion() {
	// Find the union of two slices
	result := lists.Union([]int{1, 2, 3}, []int{2, 3, 4})

	fmt.Println(result)
	// Output: [1 2 3 4]
}

func ExampleIsSorted() {
	// Check if the slice is sorted
	result := lists.IsSorted([]int{1, 2, 3}, func(a, b int) bool {
		return a < b
	})

	fmt.Println(result)
	// Output: true
}

func ExampleAllMatch() {
	// Check if all elements match the predicate
	result := lists.AllMatch([]int{1, 2, 3}, func(i int) bool {
		return i > 0
	})

	fmt.Println(result)
	// Output: true
}

func ExampleAnyMatch() {
	// Check if any element matches the predicate
	result := lists.AnyMatch([]int{1, 2, 3}, func(i int) bool {
		return i < 0
	})

	fmt.Println(result)
	// Output: false
}

func ExampleNoneMatch() {
	// Check if no elements match the predicate
	result := lists.NoneMatch([]int{1, 2, 3}, func(i int) bool {
		return i < 0
	})

	fmt.Println(result)
	// Output: true
}
