package lists_test

import (
	"fmt"

	"github.com/lvlcn-t/go-kit/lists"
)

func ExampleFilter() {
	// Filter even numbers
	result := lists.Filter([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(result)
	// Output: [2 4]
}

func ExampleFilterEmpty() {
	// Filter non-zero elements
	result := lists.FilterEmpty([]int{1, 2, 3, 0, 4})

	fmt.Println(result)
	// Output: [1 2 3 4]
}

func ExampleFilterNonEmpty() {
	// Filter zero elements
	result := lists.FilterNonEmpty([]int{1, 2, 3, 0, 4})

	fmt.Println(result)
	// Output: [0]
}

func ExampleFilterNil() {
	i := 0
	// Filter non-nil elements
	result := lists.FilterNil([]*int{nil, &i, nil})

	for _, item := range result {
		fmt.Printf("%d ", *item)
	}
	// Output: 0
}

func ExampleFilterNonNil() {
	i := 0
	// Filter nil elements
	result := lists.FilterNonNil([]*int{nil, &i, nil})

	fmt.Println(result)
	// Output: [<nil> <nil>]
}

func ExampleMatchIndex() {
	// Match index
	result := lists.MatchIndex([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(result)
	// Output: 1
}

func ExampleMatchLastIndex() {
	// Match last index
	result := lists.MatchLastIndex([]int{1, 2, 3, 4}, func(i int) bool {
		return i%2 == 0
	})

	fmt.Println(result)
	// Output: 3
}
