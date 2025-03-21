package maps

import (
	"encoding/json"
	"testing"
)

func TestHeterogeneousMap_UnmarshalJSON(t *testing.T) {
	type args struct {
		d []byte
	}
	tests := []struct {
		name     string
		args     args
		expected *Map
		wantErr  bool
	}{
		{
			name:     "simple",
			args:     args{d: []byte(`{"name":"John Doe"}`)},
			wantErr:  false,
			expected: NewFromItems("name", "John Doe"),
		},
		{
			name:     "nested",
			args:     args{d: []byte(`{"items":{"key":"value"}}`)},
			wantErr:  false,
			expected: NewFromItems("items", NewFromItems("key", "value")),
		},
		{
			name:     "number value",
			args:     args{d: []byte(`{"name":42}`)},
			wantErr:  false,
			expected: NewFromItems("name", 42),
		},
		{
			name:     "multiple items",
			args:     args{d: []byte(`{"name": "John Doe", "nested": { "gender": "non-binary" } }`)},
			wantErr:  false,
			expected: NewFromItems("name", "John Doe", "nested", NewFromItems("gender", "non-binary")),
		},
		{
			name:     "array of floats",
			args:     args{d: []byte(`{"numbers": [1.0, 2.0, 3.0]}`)},
			wantErr:  false,
			expected: NewFromItems("numbers", []any{1.0, 2.0, 3.0}),
		},
		{
			name:     "array of strings",
			args:     args{d: []byte(`{"names": ["John", "Doe"]}`)},
			wantErr:  false,
			expected: NewFromItems("names", []any{"John", "Doe"}),
		},
		{
			name:     "array of maps",
			args:     args{d: []byte(`{"sections": [ { "name": "John" }, { "name": "Doe" } ] }`)},
			expected: NewFromItems("sections", []any{NewFromItems("name", "John"), NewFromItems("name", "Doe")}),
			wantErr:  false,
		},
		{
			name:     "array of ints",
			args:     args{d: []byte(`{"numbers": [1, 2, 3]}`)},
			expected: NewFromItems("numbers", []any{1, 2, 3}),
			wantErr:  false,
		},
		{
			name:     "array of bools",
			args:     args{d: []byte(`{"bools": [true, false]}`)},
			expected: NewFromItems("bools", []any{true, false}),
			wantErr:  false,
		},
		{
			name:     "array of arrays",
			args:     args{d: []byte(`{"arrays": [[1, 2], [3, 4]]}`)},
			expected: NewFromItems("arrays", []any{[]any{1, 2}, []any{3, 4}}),
			wantErr:  false,
		},
		{
			name:    "invalid",
			args:    args{d: []byte(`{ "sections": [ { "nested": [{"name": "John"}] } ] }`)},
			wantErr: false,
			expected: NewFromItems(
				"sections", []any{
					NewFromItems(
						"nested", []any{
							NewFromItems("name", "John"),
						},
					),
				},
			),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := New()
			err := json.Unmarshal(tt.args.d, &actual)
			if tt.wantErr != (err != nil) {
				t.Fatalf("expected error %v, got %v", tt.wantErr, err)
			}
			if !actual.Equals(tt.expected) {
				actMap := actual.ToMap()
				expMap := tt.expected.ToMap()
				t.Logf("expected: %v", expMap)
				t.Logf("actual: %v", actMap)
				t.Fatalf("expected %v, got %v", tt.expected, actual)
			}
		})
	}
}

func TestHeterogeneousMap_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		m       *Map
		want    string
		wantErr bool
	}{
		{
			name:    "simple",
			m:       NewFromItems("name", "John Doe"),
			want:    `{"name":"John Doe"}`,
			wantErr: false,
		},
		{
			name:    "nested",
			m:       NewFromItems("items", NewFromItems("key", "value")),
			want:    `{"items":{"key":"value"}}`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.m.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			gotString := string(got)
			if gotString != tt.want {
				t.Errorf("MarshalJSON() got = %v, want %v", gotString, tt.want)
			}
		})
	}
}
