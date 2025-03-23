package maps

import (
	"encoding/json"
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
			if got := bestEffortJsonName(tt.in); got != tt.want {
				t.Errorf("bestEffortJsonName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToStruct(t *testing.T) {
	m := NewFromItems("a", 1, "b", 2, "c", 3)
	type testStruct struct {
		A int
		B int
		S int `champ:"c"`
	}
	s, err := ToStruct[testStruct](m)
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
	m := NewFromItems("a", nil)

	type testStruct struct {
		A string `champ:"a"`
	}

	s, err := ToStruct[testStruct](m)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("m: %v", s)
}
