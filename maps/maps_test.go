package maps

import (
	"errors"
	"reflect"
	"testing"
)

func TestFeature_Equals(t *testing.T) {
	type args struct {
		one   *Map
		other *Map
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two empty features are equal",
			args: args{
				one:   New(),
				other: New(),
			},
			want: true,
		},
		{
			name: "two features with different values are not equal",
			args: args{
				one:   NewFromItems("a", 1),
				other: NewFromItems("a", 2),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oneFeature := tt.args.one
			otherFeature := tt.args.other
			if got := oneFeature.Equals(otherFeature); got != tt.want {
				t.Errorf("Equals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeature_Set(t *testing.T) {

	f := New()

	fWithName := f.Set("name", "John")

	expected := NewFromItems("name", "John")

	isEqual := fWithName.Equals(expected)

	if !isEqual {
		t.Fatalf("have %v, want %v", fWithName, expected)
	}

}

func TestEquals(t *testing.T) {
	type args struct {
		f     *Map
		other *Map
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two empty features are equal",
			args: args{
				f:     New(),
				other: New(),
			},
			want: true,
		},
		{
			name: "two features with equal string values are equal",
			args: args{
				f:     NewFromItems("a", "1"),
				other: NewFromItems("a", "1"),
			},
			want: true,
		},
		{
			name: "two features with different string values are not equal",
			args: args{
				f:     NewFromItems("a", "1"),
				other: NewFromItems("a", "2"),
			},
			want: false,
		},
		{
			name: "two features with equal int values are equal",
			args: args{
				f:     NewFromItems("a", 1),
				other: NewFromItems("a", 1),
			},
			want: true,
		},
		{
			name: "two features with different int values are not equal",
			args: args{
				f:     NewFromItems("a", 1),
				other: NewFromItems("a", 2),
			},
			want: false,
		},
		{
			name: "two features with different keys are not equal",
			args: args{
				f:     NewFromItems("a", 1),
				other: NewFromItems("b", 1),
			},
			want: false,
		},
		{
			name: "two features with different number of keys are not equal",
			args: args{
				f:     NewFromItems("a", 1),
				other: NewFromItems("a", 1, "b", 2),
			},
			want: false,
		},
		{
			name: "two features with different types are not equal",
			args: args{
				f:     NewFromItems("a", 1),
				other: NewFromItems("a", "1"),
			},
			want: false,
		},
		{
			name: "two features with equal int32 values are equal",
			args: args{
				f:     NewFromItems("a", int32(1)),
				other: NewFromItems("a", int32(1)),
			},
			want: true,
		},
		{
			name: "two features with different int32 values are not equal",
			args: args{
				f:     NewFromItems("a", int32(1)),
				other: NewFromItems("a", int32(2)),
			},
			want: false,
		},
		{
			name: "two features with equal int64 values are equal",
			args: args{
				f:     NewFromItems("a", int64(1)),
				other: NewFromItems("a", int64(1)),
			},
			want: true,
		},
		{
			name: "two features with different int64 values are not equal",
			args: args{
				f:     NewFromItems("a", int64(1)),
				other: NewFromItems("a", int64(2)),
			},
			want: false,
		},
		{
			name: "many fields in different order are equal",
			args: args{
				f:     NewFromItems("a", 1, "b", 2, "c", 3, "d", 4, "e", 5),
				other: NewFromItems("e", 5, "d", 4, "c", 3, "b", 2, "a", 1),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.args.f.Equals(tt.args.other); got != tt.want {
				diff, _ := tt.args.f.Diff(tt.args.other)
				t.Errorf("Equals() = %v, want %v", got, tt.want)
				t.Errorf("Diff() = %v", diff)
			}
		})
	}
}

func TestNewFromItems(t *testing.T) {
	tests := []struct {
		name    string
		items   []any
		want    *Map
		wantErr error
	}{
		{
			name:    "empty items creates an empty feature",
			items:   []any{},
			want:    New(),
			wantErr: nil,
		},
		{
			name:    "valid items creates a feature",
			items:   []any{"a", 1, "b", 2},
			want:    NewFromItems("a", 1, "b", 2),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fromItems := NewFromItems(tt.items...)
			if !fromItems.Equals(tt.want) {
				t.Errorf("got %v, want %v", fromItems, tt.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	type args struct {
		one   *Map
		other *Map
	}
	tests := []struct {
		name            string
		args            args
		expectedHasDiff bool
		wantDiff        *Map
	}{
		{
			name: "two empty features are equal",
			args: args{
				one:   New(),
				other: New(),
			},
			expectedHasDiff: false,
			wantDiff:        New(),
		},
		{
			name: "two features with equal values are equal",
			args: args{
				one:   NewFromItems("a", 1),
				other: NewFromItems("a", 1),
			},
			expectedHasDiff: false,
			wantDiff:        New(),
		},
		{
			name: "one feature has extra key",
			args: args{
				one:   NewFromItems("a", 1),
				other: NewFromItems("a", 1, "b", 2),
			},
			expectedHasDiff: true,
			wantDiff:        NewFromItems("b", 2),
		},
		{
			name: "nested features are equal",
			args: args{
				one:   NewFromItems("a", NewFromItems("b", 1)),
				other: NewFromItems("a", NewFromItems("b", 1)),
			},
			expectedHasDiff: false,
			wantDiff:        New(),
		},
		{
			name: "nested features are not equal",
			args: args{
				one:   NewFromItems("a", NewFromItems("b", 1)),
				other: NewFromItems("a", NewFromItems("b", 2)),
			},
			expectedHasDiff: true,
			wantDiff:        NewFromItems("a", NewFromItems("b", 2)),
		},
		{
			name: "empty feature and feature with key are not equal",
			args: args{
				one:   New(),
				other: NewFromItems("a", 1),
			},
			expectedHasDiff: true,
			wantDiff:        NewFromItems("a", 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			one := tt.args.one
			other := tt.args.other

			diff, err := one.Diff(other)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tt.expectedHasDiff && !diff.Equals(tt.wantDiff) {
				t.Errorf("expected equality: %v", tt.expectedHasDiff)
			}
			if !diff.Equals(tt.wantDiff) {
				t.Errorf("expected diff: %v, got: %v", tt.wantDiff, diff)
			}
		})
	}
}

func TestDiffSlice(t *testing.T) {
	type args struct {
		one   []any
		other []any
	}
	tests := []struct {
		name     string
		args     args
		wantDiff bool
	}{
		{
			name: "two empty slices are equal",
			args: args{
				one:   []any{},
				other: []any{},
			},
			wantDiff: false,
		},
		{
			name: "two slices with equal values are equal",
			args: args{
				one:   []any{1},
				other: []any{1},
			},
			wantDiff: false,
		},
		{
			name: "two slices with different values are not equal",
			args: args{
				one:   []any{1},
				other: []any{2},
			},
			wantDiff: true,
		},
		{
			name: "nested slices with equal values are equal",
			args: args{
				one:   []any{[]any{1}},
				other: []any{[]any{1}},
			},
			wantDiff: false,
		},
		{
			name: "nested slices with different values are not equal",
			args: args{
				one:   []any{[]any{1}},
				other: []any{[]any{2}},
			},
			wantDiff: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, hasDiff := DiffSlice(tt.args.one, tt.args.other)
			if hasDiff != tt.wantDiff {
				t.Errorf("expected diff: %v, got: %v", tt.wantDiff, hasDiff)
			}
		})
	}
}

func TestDiffMapNoNested(t *testing.T) {
	tests := []struct {
		name          string
		first         *Map
		second        *Map
		diff          *Map
		expectedError error
	}{
		{
			name:          "value change",
			first:         NewFromItems("a", 1),
			second:        NewFromItems("a", 2),
			diff:          NewFromItems("a", 2),
			expectedError: nil,
		},
		{
			name:          "new key",
			first:         NewFromItems("a", 1),
			second:        NewFromItems("a", 1, "b", 2),
			diff:          NewFromItems("b", 2),
			expectedError: nil,
		},
		{
			name:          "key removed",
			first:         NewFromItems("a", 1, "b", 2),
			second:        NewFromItems("a", 1),
			diff:          NewFromItems("b", nil),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := tt.first.Diff(tt.second)
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if tt.expectedError != nil {
				return
			}
			if !diff.Equals(tt.diff) {
				t.Fatalf("expected diff: %v, got: %v", tt.diff, diff)
			}
			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestDiffMapNested(t *testing.T) {
	tests := []struct {
		name          string
		first         *Map
		second        *Map
		diff          *Map
		expectedError error
	}{
		{
			name:          "value change",
			first:         NewFromItems("a", NewFromItems("b", 1)),
			second:        NewFromItems("a", NewFromItems("b", 2)),
			diff:          NewFromItems("a", NewFromItems("b", 2)),
			expectedError: nil,
		},
		{
			name:          "new key",
			first:         NewFromItems("a", NewFromItems("b", 1)),
			second:        NewFromItems("a", NewFromItems("b", 1, "c", 2)),
			diff:          NewFromItems("a", NewFromItems("c", 2)),
			expectedError: nil,
		},
		{
			name:          "key removed",
			first:         NewFromItems("a", NewFromItems("b", 1, "c", 2)),
			second:        NewFromItems("a", NewFromItems("b", 1)),
			diff:          NewFromItems("a", NewFromItems("c", nil)),
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff, err := tt.first.Diff(tt.second)

			if !errors.Is(err, tt.expectedError) {
				t.Fatalf("expected error: %v, got: %v", tt.expectedError, err)
			}
			if tt.expectedError != nil {
				return
			}
			if !diff.Equals(tt.diff) {
				t.Errorf("expected diff: %v, got: %v", tt.diff, diff)
			}
			if !errors.Is(err, tt.expectedError) {
				t.Errorf("expected error: %v, got: %v", tt.expectedError, err)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	type args struct {
		one   []string
		other []string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "two empty slices have no intersection",
			args: args{
				one:   []string{},
				other: []string{},
			},
			want: []string{},
		},
		{
			name: "two slices with equal values have intersection",
			args: args{
				one:   []string{"a"},
				other: []string{"a"},
			},
			want: []string{"a"},
		},
		{
			name: "two slices with different values have no intersection",
			args: args{
				one:   []string{"a"},
				other: []string{"b"},
			},
			want: []string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Intersection(tt.args.one, tt.args.other)
			if len(got) != len(tt.want) {
				t.Fatalf("have %v, want %v", got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("have %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestMerge(t *testing.T) {
	type args struct {
		current *Map
		diff    *Map
	}
	tests := []struct {
		name string
		args args
		want *Map
	}{
		{
			name: "empty diff",
			args: args{
				current: New(),
				diff:    New(),
			},
			want: New(),
		},
		{
			name: "empty current, simple scalar diff",
			args: args{
				current: New(),
				diff:    NewFromItems("a", 1),
			},
			want: NewFromItems("a", 1),
		},
		{
			name: "empty current, simple nested diff",
			args: args{
				current: New(),
				diff:    NewFromItems("a", NewFromItems("b", 1)),
			},
			want: NewFromItems("a", NewFromItems("b", 1)),
		},
		{
			name: "nested current, nested diff without conflict",
			args: args{
				current: NewFromItems("a", NewFromItems("b", 1)),
				diff:    NewFromItems("a", NewFromItems("c", 2)),
			},
			want: NewFromItems("a", NewFromItems("b", 1, "c", 2)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.current.Merge(tt.args.diff)
			if !got.Equals(tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHavePathInCommon(t *testing.T) {
	type args struct {
		a *Map
		b *Map
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "two empty features have no path in common",
			args: args{
				a: New(),
				b: New(),
			},
			want: false,
		},
		{
			name: "two features with equal values have path in common",
			args: args{
				a: NewFromItems("a", 1),
				b: NewFromItems("a", 1),
			},
			want: true,
		},
		{
			name: "two features with different values have path in common",
			args: args{
				a: NewFromItems("a", 1),
				b: NewFromItems("a", 2),
			},
			want: true,
		},
		{
			name: "two features with nested features have path in common",
			args: args{
				a: NewFromItems("a", NewFromItems("b", 1)),
				b: NewFromItems("a", NewFromItems("b", 2)),
			},
			want: true,
		},
		{
			name: "two features with nested features have no path in common",
			args: args{
				a: NewFromItems("a", NewFromItems("b", 1)),
				b: NewFromItems("a", NewFromItems("c", 2)),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HavePathInCommon(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("HavePathInCommon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInformationPaths(t *testing.T) {
	f := NewFromItems("a", 1, "b", NewFromItems("c", 2))
	paths := InformationPaths(f)
	expected := []string{"a", "b.c"}

	if _, hasDiff := DiffSlice(paths, expected); hasDiff {
		t.Fatalf("have %v, want %v", paths, expected)
	}
}

func TestRefToLookup(t *testing.T) {
	type args struct {
		ref *Map
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty ref",
			args: args{
				ref: New(),
			},
			want: []string{},
		},
		{
			name: "simple ref",
			args: args{
				ref: NewFromItems("$ref", "#/property"),
			},
			want: []string{"property"},
		},
		{
			name: "nested ref",
			args: args{
				ref: NewFromItems("$ref", "#/property/nested"),
			},
			want: []string{"property", "nested"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := RefToLookup(tt.args.ref); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RefToLookup() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name     string
		f        *Map
		key      any
		expected any
	}{
		{
			name:     "simple key",
			f:        NewFromItems("a", 1),
			key:      "a",
			expected: 1,
		},
		{
			name:     "nested key with string slice key",
			f:        NewFromItems("a", NewFromItems("b", 1)),
			key:      []string{"a", "b"},
			expected: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, ok := tt.f.Get(tt.key); !ok || !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("Get() = %v, want %v", got, tt.expected)
			}
		})
	}
}
