package jsonchamp

import (
	"encoding/json"
	"reflect"
	"testing"
)

type testFeatureAttributes struct {
	NestedAttribute struct {
		Integer int `json:"integer"`
	} `json:"nestedAttribute"`
}

type testFeature struct {
	Attributes testFeatureAttributes `json:"attributes"`
}

func TestConvertToStruct(t *testing.T) {
	t.Parallel()

	//nolint: lll
	const data = `{"fieldExcludedFromJson": "fieldExcludedFromJson",    "fieldIncludedInJson": "fieldIncludedInJson",    "attributes": {       "array": [1, 2, 3],       "attributeIncludedInJson": "attributeIncludedInJson",       "attributeThatIsInt64": 123453153,       "attributeThatIsMaxInt64": 9223372036854775807,       "attributeOutsideJson": "attributeOutsideJson",       "anotherAttributeOutsideJson": "anotherAttributeOutsideJson",       "nestedAttribute": {          "integer": 123       }    },    "geometry": {       "x": 129494214.0,       "y": 12412421.1    }}`

	var m *Map
	if err := json.Unmarshal([]byte(data), &m); err != nil {
		t.Fatal(err)
	}

	f, err := To[testFeature](m)
	if err != nil {
		t.Fatal(err)
	}

	f.Attributes.NestedAttribute.Integer = 666

	featureMap, err := From(f)
	if err != nil {
		t.Fatal(err)
	}

	newOut := m.Merge(featureMap)

	t.Logf("m: %v", newOut.ToMap())
}

func Test_bestEffortJsonName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"single", "A", "a"},
		{"two", "AB", "a_b"},
		{"three", "ABC", "a_b_c"},
		{"words", "HelloWorld", "hello_world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := bestEffortJSONName(tt.in); got != tt.want {
				t.Errorf("bestEffortJSONName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToStruct(t *testing.T) {
	t.Parallel()

	m := NewFromItems("a", 1, "b", 2, "c", 3)

	type testStruct struct {
		A int
		B int
		S int `champ:"c"`
	}

	var s testStruct

	err := ToStruct(m, &s)
	if err != nil {
		t.Fatal(err)
	}

	if s.A != 1 {
		t.Errorf("expected 1, got %d", s.A)
	}

	if s.B != 2 {
		t.Errorf("expected 2, got %d", s.B)
	}

	if s.S != 3 {
		t.Errorf("expected 3, got %d", s.S)
	}
}

func TestToStructWithMoreCoverage(t *testing.T) {
	t.Parallel()

	m := NewFromItems("a", nil)

	type testStruct struct {
		A string `champ:"a"`
	}

	var s testStruct

	err := ToStruct(m, &s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("m: %v", s)
}

func TestToNativeMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   *Map
		want map[string]any
	}{
		{
			name: "simple",
			in:   NewFromItems("a", 1, "b", 2, "c", 3),
			want: map[string]any{"a": 1, "b": 2, "c": 3},
		},
		{
			name: "nested",
			in:   NewFromItems("a", 1, "b", NewFromItems("c", 3)),
			want: map[string]any{"a": 1, "b": map[string]any{"c": 3}},
		},
		{
			name: "slice of map",
			in:   NewFromItems("a", 1, "b", []any{NewFromItems("c", 3), NewFromItems("d", 4)}),
			want: map[string]any{"a": 1, "b": []map[string]any{{"c": 3}, {"d": 4}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := ToNativeMap(tt.in)
			want := normalizeNativeMap(tt.want)

			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ToNativeMap() = %v, want %v", got, want)
			}
		})
	}
}

func TestFromNativeMap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   map[string]any
		want *Map
	}{
		{
			name: "simple",
			in:   map[string]any{"a": 1, "b": 2, "c": 3},
			want: NewFromItems("a", 1, "b", 2, "c", 3),
		},
		{
			name: "nested",
			in:   map[string]any{"a": 1, "b": map[string]any{"c": 3}},
			want: NewFromItems("a", 1, "b", NewFromItems("c", 3)),
		},
		{
			name: "slice of map",
			in:   map[string]any{"a": 1, "b": []map[string]any{{"c": 3}, {"d": 4}}},
			want: NewFromItems("a", 1, "b", []any{NewFromItems("c", 3), NewFromItems("d", 4)}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := FromNativeMap(tt.in)

			if diff, _ := got.Diff(tt.want); len(diff.Keys()) > 0 {
				out, _ := json.MarshalIndent(diff, "", "  ")
				t.Fatalf("%s", out)
			}
		})
	}
}
