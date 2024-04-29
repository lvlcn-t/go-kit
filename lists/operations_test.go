package lists

import (
	"reflect"
	"testing"
)

func TestOperations_Apply(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		f     func(int) int
		want  []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3},
			f: func(i int) int {
				return i * i
			},
			want: []int{1, 4, 9},
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) int {
				return i * i
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Apply(tt.slice, tt.f)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Map() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Reduce(t *testing.T) {
	tests := []struct {
		name    string
		slice   []int
		initial int
		f       func(int, int) int
		want    int
	}{
		{
			name:    "success",
			slice:   []int{1, 2, 3},
			initial: 0,
			f: func(acc, item int) int {
				return acc + item
			},
			want: 6,
		},
		{
			name:    "empty slice",
			slice:   []int{},
			initial: 0,
			f: func(acc, item int) int {
				return acc + item
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Reduce(tt.slice, tt.initial, tt.f)
			if got != tt.want {
				t.Errorf("Reduce() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_LastIndexOf(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		value int
		want  int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 2},
			value: 2,
			want:  3,
		},
		{
			name:  "empty slice",
			slice: []int{},
			value: 2,
			want:  -1,
		},
		{
			name:  "no match",
			slice: []int{1, 3, 5},
			value: 2,
			want:  -1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := LastIndexOf(tt.slice, tt.value)
			if got != tt.want {
				t.Errorf("LastIndexOf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Count(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		f     func(int) bool
		want  int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: 2,
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Count(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("Count() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Distinct(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		want  []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 2, 1},
			want:  []int{1, 2, 3},
		},
		{
			name:  "empty slice",
			slice: []int{},
			want:  []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Distinct(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Distinct() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Partition(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		f     func(int) bool
		wantT []int
		wantF []int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			wantT: []int{2, 4},
			wantF: []int{1, 3},
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			wantT: []int{},
			wantF: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trueSlice, falseSlice := Partition(tt.slice, tt.f)
			if !reflect.DeepEqual(trueSlice, tt.wantT) {
				t.Errorf("Partition() trueSlice = %v, want %v", trueSlice, tt.wantT)
			}
			if !reflect.DeepEqual(falseSlice, tt.wantF) {
				t.Errorf("Partition() falseSlice = %v, want %v", falseSlice, tt.wantF)
			}
		})
	}
}

func TestOperations_Permutations(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		// We can't compare the result with a specific value because the permutations are not deterministic.
		wantLength int
	}{
		{
			name:       "success",
			slice:      []int{1, 2, 3},
			wantLength: 6,
		},
		{
			name:       "empty slice",
			slice:      []int{},
			wantLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Permutations(tt.slice)
			if len(got) != tt.wantLength {
				t.Errorf("Permutations() = %v; length = %d, want length %d", got, len(got), tt.wantLength)
			}
		})
	}
}

func TestOperations_Combinations(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		n     int
		// We can't compare the result with a specific value because the combinations are not deterministic.
		wantLength int
	}{
		{
			name:       "success",
			slice:      []int{1, 2, 3},
			n:          2,
			wantLength: 3,
		},
		{
			name:       "empty slice",
			slice:      []int{},
			n:          2,
			wantLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Combinations(tt.slice, tt.n)
			if len(got) != tt.wantLength {
				t.Errorf("Combinations() = %v; length = %d, want length %d", got, len(got), tt.wantLength)
			}
		})
	}
}

func TestOperations_Shuffle(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		// We can't compare the result with a specific value because the shuffle is not deterministic.
		wantLength int
	}{
		{
			name:       "success",
			slice:      []int{1, 2, 3, 4},
			wantLength: 4,
		},
		{
			name:       "empty slice",
			slice:      []int{},
			wantLength: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Shuffle(tt.slice)
			if len(got) != tt.wantLength {
				t.Errorf("Shuffle() = %v; length = %d, want length %d", got, len(got), tt.wantLength)
			}
		})
	}
}

func TestOperations_Zip(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		other []string
		want  []Pair[int, string]
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3},
			other: []string{"a", "b", "c"},
			want: []Pair[int, string]{
				{First: 1, Second: "a"},
				{First: 2, Second: "b"},
				{First: 3, Second: "c"},
			},
		},
		{
			name:  "empty slice",
			slice: []int{},
			other: []string{"a", "b", "c"},
			want:  []Pair[int, string]{},
		},
		{
			name:  "empty other",
			slice: []int{1, 2, 3},
			other: []string{},
			want:  []Pair[int, string]{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Zip(tt.slice, tt.other)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Zip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Unzip(t *testing.T) {
	tests := []struct {
		name  string
		pairs []Pair[int, string]
		want1 []int
		want2 []string
	}{
		{
			name: "success",
			pairs: []Pair[int, string]{
				{First: 1, Second: "a"},
				{First: 2, Second: "b"},
				{First: 3, Second: "c"},
			},
			want1: []int{1, 2, 3},
			want2: []string{"a", "b", "c"},
		},
		{
			name:  "empty slice",
			pairs: []Pair[int, string]{},
			want1: []int{},
			want2: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got1, got2 := Unzip(tt.pairs)
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Unzip() got1 = %v, want %v", got1, tt.want1)
			}
			if !reflect.DeepEqual(got2, tt.want2) {
				t.Errorf("Unzip() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

func TestOperations_Chunk(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		size  int
		want  [][]int
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4, 5},
			size:  2,
			want: [][]int{
				{1, 2},
				{3, 4},
				{5},
			},
		},
		{
			name:  "empty slice",
			slice: []int{},
			size:  2,
			want:  [][]int{},
		},
		{
			name:  "size greater than slice",
			slice: []int{1, 2, 3},
			size:  5,
			want: [][]int{
				{1, 2, 3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Chunk(tt.slice, tt.size)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Chunk() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Flatten(t *testing.T) {
	tests := []struct {
		name  string
		slice [][]int
		want  []int
	}{
		{
			name: "success",
			slice: [][]int{
				{1, 2},
				{3, 4},
				{5},
			},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name:  "empty slice",
			slice: [][]int{},
			want:  []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Flatten(tt.slice)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Flatten() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Intersect(t *testing.T) {
	tests := []struct {
		name   string
		slices [][]int
		want   []int
	}{
		{
			name: "success",
			slices: [][]int{
				{1, 2, 3},
				{2, 3, 4},
				{3, 4, 5},
			},
			want: []int{3},
		},
		{
			name:   "empty slice",
			slices: [][]int{},
			want:   []int{},
		},
		{
			name: "no common element",
			slices: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			want: []int{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Intersect(tt.slices...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Intersect() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Difference(t *testing.T) {
	tests := []struct {
		name   string
		slices [][]int
		want   []int
	}{
		{
			name: "success",
			slices: [][]int{
				{1, 2, 3},
				{2, 3, 4},
				{3, 4, 5},
			},
			want: []int{1},
		},
		{
			name:   "empty slice",
			slices: [][]int{},
			want:   []int{},
		},
		{
			name: "no common element",
			slices: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			want: []int{1, 2, 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Difference(tt.slices...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Difference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_Union(t *testing.T) {
	tests := []struct {
		name   string
		slices [][]int
		want   []int
	}{
		{
			name: "success",
			slices: [][]int{
				{1, 2, 3},
				{2, 3, 4},
				{3, 4, 5},
			},
			want: []int{1, 2, 3, 4, 5},
		},
		{
			name:   "empty slice",
			slices: [][]int{},
			want:   []int{},
		},
		{
			name: "no common element",
			slices: [][]int{
				{1, 2, 3},
				{4, 5, 6},
				{7, 8, 9},
			},
			want: []int{1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Union(tt.slices...)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Union() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_IsSorted(t *testing.T) {
	tests := []struct {
		name  string
		slice []int
		f     func(int, int) bool
		want  bool
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(a, b int) bool {
				return a < b
			},
			want: true,
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(a, b int) bool {
				return a < b
			},
			want: true,
		},
		{
			name:  "not sorted",
			slice: []int{1, 3, 2},
			f: func(a, b int) bool {
				return a < b
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSorted(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("IsSorted() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_AllMatch(t *testing.T) { //nolint:dupl // generic functions cannot be stored in variables
	tests := []struct {
		name  string
		slice []int
		f     func(int) bool
		want  bool
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: false,
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: true,
		},
		{
			name:  "all match",
			slice: []int{2, 4, 6},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AllMatch(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("AllMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_AnyMatch(t *testing.T) { //nolint:dupl // generic functions cannot be stored in variables
	tests := []struct {
		name  string
		slice []int
		f     func(int) bool
		want  bool
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: true,
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: false,
		},
		{
			name:  "no match",
			slice: []int{1, 3, 5},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := AnyMatch(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("AnyMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperations_NoneMatch(t *testing.T) { //nolint:dupl // generic functions cannot be stored in variables
	tests := []struct {
		name  string
		slice []int
		f     func(int) bool
		want  bool
	}{
		{
			name:  "success",
			slice: []int{1, 2, 3, 4},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: false,
		},
		{
			name:  "empty slice",
			slice: []int{},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: true,
		},
		{
			name:  "no match",
			slice: []int{1, 3, 5},
			f: func(i int) bool {
				return i%2 == 0
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NoneMatch(tt.slice, tt.f)
			if got != tt.want {
				t.Errorf("NoneMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
