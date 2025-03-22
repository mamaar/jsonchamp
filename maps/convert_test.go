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
