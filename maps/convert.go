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

func ToStruct(m *Map, out any) error {
	structValue := reflect.ValueOf(out).Elem()

	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}

	structType := structValue.Type()
	if structType.Kind() != reflect.Struct {
		return fmt.Errorf("expected struct, got %v", structType.Kind())
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

		// If the map value is nil, we set the field to the zero value of the field type.
		if mapVal == nil {
			mapVal = reflect.Zero(fieldType.Type).Interface()
		}

		switch fieldType.Type.Kind() {
		case reflect.Int:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(int(0))))
		case reflect.Int8:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(int8(0))))
		case reflect.Int16:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(int16(0))))
		case reflect.Int32:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(int32(0))))
		case reflect.Int64:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(int64(0))))
		case reflect.Uint:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(uint(0))))
		case reflect.Uint8:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(uint8(0))))
		case reflect.Uint16:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(uint16(0))))
		case reflect.Uint32:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(uint32(0))))
		case reflect.Uint64:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(uint64(0))))
		case reflect.Float32:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(float32(0))))
		case reflect.Float64:
			structValue.Field(i).Set(reflect.ValueOf(mapVal).Convert(reflect.TypeOf(float64(0))))
		case reflect.String:
			structValue.Field(i).SetString(mapVal.(string))
		case reflect.Struct:
			err := ToStruct(mapVal.(*Map), structValue.Field(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to convert map to struct: %w", err)
			}
		case reflect.Bool:
			structValue.Field(i).SetBool(mapVal.(bool))
		case reflect.Slice:
			slice := reflect.ValueOf(mapVal)
			if slice.Len() == 0 {
				structValue.Field(i).Set(reflect.Zero(fieldType.Type))
				continue
			}
			structValue.Field(i).Set(reflect.ValueOf(mapVal))
		case reflect.Map:
			return fmt.Errorf("map fields are not supported. use a struct instead")
		case reflect.Ptr:
			return fmt.Errorf("pointer fields are not supported")
		default:
			return fmt.Errorf("unsupported type: %v", fieldType.Type.Kind())
		}
	}

	return nil
}

func FromNativeMap(in map[string]any) *Map {
	res := New()
	for k, v := range in {
		if subMap, isMap := v.(map[string]any); isMap {
			res = res.Set(k, FromNativeMap(subMap))
		} else {
			res = res.Set(k, v)
		}
	}
	return res
}

func ToNativeMap(in *Map) map[string]any {
	res := make(map[string]any)
	for _, k := range in.Keys() {
		v, _ := in.Get(k)
		if subMap, isMap := v.(*Map); isMap {
			res[k] = ToNativeMap(subMap)
		} else {
			res[k] = v
		}
	}
	return res
}
