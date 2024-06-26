package lists

import (
	"reflect"
	"testing"
)

func TestFilters_Filter(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		f     Predicate[int]
		want  []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: []int{2, 4},
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: []int{},
		},
		{
			name:  "no match",
			slice: []int{1, 3, 5},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Filter(tt.slice, tt.f)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilters_FilterEmpty(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		want  []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 0, 4},
			want:  []int{1, 2, 3, 4},
		},
		{
			name:  "empty slice",
			slice: []int{},
			want:  []int{},
		},
		{
			name:  "no match",
			slice: []int{0, 0, 0},
			want:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterEmpty(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilters_FilterNil(t *testing.T) {
	tests := []struct {
		name  string
		slice []*int
		want  []*int
	}{
		{
			name:  "success",
			slice: []*int{nil, new(int), nil},
			want:  []*int{new(int)},
		},
		{
			name:  "empty slice",
			slice: []*int{},
			want:  []*int{},
		},
		{
			name:  "no match",
			slice: []*int{nil, nil, nil},
			want:  []*int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterNil(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilters_FilterNonEmpty(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		want  []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 0, 4},
			want:  []int{0},
		},
		{
			name:  "empty slice",
			slice: []int{},
			want:  []int{},
		},
		{
			name:  "no match",
			slice: []int{1, 2, 3},
			want:  []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterNonEmpty(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterNonEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilters_FilterNonNil(t *testing.T) {
	tests := []struct {
		name  string
		slice []*int
		want  []*int
	}{
		{
			name:  "success",
			slice: []*int{nil, new(int), nil},
			want:  []*int{nil, nil},
		},
		{
			name:  "empty slice",
			slice: []*int{},
			want:  []*int{},
		},
		{
			name:  "no match",
			slice: []*int{new(int), new(int), new(int)},
			want:  []*int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterNonNil(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilterNonNil() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilters_MatchIndex_MatchLastIndex(t *testing.T) {
	type testFunc[T any] func(slice []T, f Predicate[T]) int

	tests := []struct {
		name  string
		slice []int
		f     Predicate[int]
		fun   testFunc[int]
		want  int
	}{
		{
			name:  "success - MatchIndex",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchIndex[int],
			want: 1,
		},
		{
			name:  "empty slice - MatchIndex",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchIndex[int],
			want: -1,
		},
		{
			name:  "no match - MatchIndex",
			slice: []int{1, 3, 5},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchIndex[int],
			want: -1,
		},
		{
			name:  "success - MatchLastIndex",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchLastIndex[int],
			want: 3,
		},
		{
			name:  "empty slice - MatchLastIndex",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchLastIndex[int],
			want: -1,
		},
		{
			name:  "no match - MatchLastIndex",
			slice: []int{1, 3, 5},
			f: func(i int) bool {
				return i%2 == 0
			},
			fun:  MatchLastIndex[int],
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fun(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("%s() = %v, want %v", reflect.TypeOf(tt.fun).Name(), got, tt.want)
			}
		})
	}
}
