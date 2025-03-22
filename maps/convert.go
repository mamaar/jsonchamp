package maps

import (
	"encoding/json"
	"fmt"
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
