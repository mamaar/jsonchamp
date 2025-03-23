package maps

import (
	"encoding/json"
	"fmt"
	"reflect"
	"unicode"
)

// To converts a Map to a struct by marshalling into JSON and then unmarshalling into the struct.
func To[T any](m *Map) (T, error) {
	var f T

	j, err := json.Marshal(m)
	if err != nil {
		return f, fmt.Errorf("failed to convert map to struct: %w", err)
	}

	if err := json.Unmarshal(j, &f); err != nil {
		return f, fmt.Errorf("failed to convert map to struct: %w", err)
	}

	return f, nil
}

// From converts a struct to a Map by marshalling into JSON and then unmarshalling into the Map.
func From[T any](f T) (*Map, error) {
	j, err := json.Marshal(f)
	if err != nil {
		return nil, fmt.Errorf("failed to convert struct to map: %w", err)
	}

	m := New()
	if err := json.Unmarshal(j, &m); err != nil {
		return nil, fmt.Errorf("failed to convert struct to map: %w", err)
	}

	return m, nil
}

// bestEffortJsonName converts a string to a string that can be used as a JSON key.
// It is best effort because we don't know the naming convention the user wants to use.
// We use snake_case as a best effort.
func bestEffortJsonName(n string) string {
	snakeCased := ""
	for i, r := range n {
		if i > 0 && r >= 'A' && r <= 'Z' {
			snakeCased += "_"
		}
		snakeCased += string(unicode.ToLower(r))
	}
	return snakeCased
}

func ToStruct[T any](m *Map) (T, error) {
	var f T

	structType := reflect.TypeOf(f)
	structValue := reflect.ValueOf(&f).Elem()
	if structType.Kind() != reflect.Struct {
		return f, fmt.Errorf("expected struct, got %v", structType.Kind())
	}

	numFields := structValue.NumField()
	for i := 0; i < numFields; i++ {
		fieldType := structType.Field(i)
		champName := fieldType.Tag.Get("champ")
		if champName == "" {
			champName = bestEffortJsonName(fieldType.Name)
		}

		mapVal, ok := m.Get(champName)
		if !ok {
			continue
		}
		fieldValue := structValue.Field(i)
		fieldValue.Set(reflect.ValueOf(mapVal))
	}

	return f, nil
}
