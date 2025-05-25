package jsonchamp

import (
	"reflect"
	"testing"
)

func TestMapSet(t *testing.T) {
	t.Parallel()

	f := New()

	fWithName := f.Set("name", "John")

	expected := NewFromItems("name", "John")

	isEqual := fWithName.Equals(expected)

	if !isEqual {
		t.Fatalf("have %v, want %v", fWithName, expected)
	}
}

func TestMapEquals(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			if got := tt.args.f.Equals(tt.args.other); got != tt.want {
				diff := tt.args.f.Diff(tt.args.other)
				t.Errorf("Equals() = %v, want %v", got, tt.want)
				t.Errorf("Diff() = %v", diff)
			}
		})
	}
}

func TestNewFromItems(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			fromItems := NewFromItems(tt.items...)
			if !fromItems.Equals(tt.want) {
				t.Errorf("got %v, want %v", fromItems, tt.want)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			one := tt.args.one
			other := tt.args.other

			diff := one.Diff(other)

			if tt.expectedHasDiff && !diff.Equals(tt.wantDiff) {
				t.Errorf("expected equality: %v", tt.expectedHasDiff)
			}

			if !diff.Equals(tt.wantDiff) {
				t.Errorf("expected diff: %v, got: %v", tt.wantDiff, diff)
			}
		})
	}
}

func TestDiffMapNoNested(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			diff := tt.first.Diff(tt.second)

			if tt.expectedError != nil {
				return
			}

			if !diff.Equals(tt.diff) {
				t.Fatalf("expected diff: %v, got: %v", tt.diff, diff)
			}
		})
	}
}

func TestDiffMapNested(t *testing.T) {
	t.Parallel()

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
			t.Parallel()
			diff := tt.first.Diff(tt.second)

			if tt.expectedError != nil {
				return
			}

			if !diff.Equals(tt.diff) {
				t.Errorf("expected diff: %v, got: %v", tt.diff, diff)
			}
		})
	}
}

func TestIntersection(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			got := intersection(tt.args.one, tt.args.other)
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
	t.Parallel()

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
			t.Parallel()

			got := tt.args.current.Merge(tt.args.diff)
			if !got.Equals(tt.want) {
				t.Errorf("Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHavePathInCommon(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			if got := havePathInCommon(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("havePathInCommon() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGet(t *testing.T) {
	t.Parallel()

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
			t.Parallel()

			got, ok := tt.f.Get(tt.key)
			expected := normalizeValue(tt.expected)

			if !ok || !reflect.DeepEqual(got, expected) {
				t.Errorf("get() = %v, want %v", got, expected)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func() *Map
		key         string
		expectedMap *Map
	}{
		{
			name: "delete from empty map",
			setup: func() *Map {
				return New()
			},
			key:         "key",
			expectedMap: New(),
		},
		{
			name: "delete from map with one value",
			setup: func() *Map {
				return NewFromItems("key", 123)
			},
			key:         "key",
			expectedMap: New(),
		},
		{
			name: "delete from map with multiple values",
			setup: func() *Map {
				return NewFromItems("key", 123, "key_2", 456)
			},
			key:         "key",
			expectedMap: NewFromItems("key_2", 456),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			m := testCase.setup()
			m, _ = m.Delete(testCase.key)

			if !m.Equals(testCase.expectedMap) {
				t.Errorf("delete(%s) = %v; want %v", testCase.key, m, testCase.expectedMap)
			}
		})
	}
}
