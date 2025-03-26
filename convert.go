package jsonchamp

import (
	"encoding/json"
	"errors"
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

// bestEffortJSONName converts a string to a string that can be used as a JSON key.
// It is best effort because we don't know the naming convention the user wants to use.
// We use snake_case as a best effort.
func bestEffortJSONName(n string) string {
	snakeCased := ""

	for i, r := range n {
		if i > 0 && r >= 'A' && r <= 'Z' {
			snakeCased += "_"
		}

		snakeCased += string(unicode.ToLower(r))
	}

	return snakeCased
}

// ToStruct converts a Map to a struct without using JSON marshalling.
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
	for i := range numFields {
		fieldType := structType.Field(i)
		champName := fieldType.Tag.Get("champ")

		if champName == "" {
			champName = bestEffortJSONName(fieldType.Name)
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
			stringed, ok := mapVal.(string)
			if !ok {
				return fmt.Errorf("expected field %s to be a string, got %T", champName, mapVal)
			}

			structValue.Field(i).SetString(stringed)
		case reflect.Struct:
			mapValTyped, ok := mapVal.(*Map)
			if !ok {
				return fmt.Errorf("expected field %s to be a map, got %T", champName, mapVal)
			}

			err := ToStruct(mapValTyped, structValue.Field(i).Addr().Interface())
			if err != nil {
				return fmt.Errorf("failed to convert map to struct: %w", err)
			}
		case reflect.Bool:
			booled, ok := mapVal.(bool)
			if !ok {
				return fmt.Errorf("expected field %s to be a bool, got %T", champName, mapVal)
			}

			structValue.Field(i).SetBool(booled)
		case reflect.Slice:
			slice := reflect.ValueOf(mapVal)
			if slice.Len() == 0 {
				structValue.Field(i).Set(reflect.Zero(fieldType.Type))

				continue
			}

			structValue.Field(i).Set(reflect.ValueOf(mapVal))
		case reflect.Map:
			return errors.New("map fields are not supported. use a struct instead")
		case reflect.Ptr:
			return errors.New("pointer fields are not supported")
		default:
			return fmt.Errorf("unsupported type: %v", fieldType.Type.Kind())
		}
	}

	return nil
}

// FromNativeMap converts a native map to a jsonchamp Map.
func FromNativeMap(in map[string]any) *Map {
	res := New()

	for k, v := range in {
		switch t := v.(type) {
		case map[string]any:
			res = res.Set(k, FromNativeMap(t))
		case []map[string]any:
			var arr []*Map
			for _, m := range t {
				arr = append(arr, FromNativeMap(m))
			}

			res = res.Set(k, arr)
		default:
			res = res.Set(k, v)
		}
	}

	return res
}

func toNativeSlice(in []any) []any {
	var res []any

	for _, v := range in {
		switch t := v.(type) {
		case map[string]any:
			res = append(res, normalizeNativeMap(t))
		case []any:
			res = append(res, toNativeSlice(t))
		case *Map:
			res = append(res, ToNativeMap(t))
		default:
			res = append(res, v)
		}
	}

	return res
}

// ToNativeMap converts a jsonchamp Map to a native map.
func ToNativeMap(in *Map) map[string]any {
	res := make(map[string]any)

	for _, k := range in.Keys() {
		v, _ := in.Get(k)
		switch reflect.TypeOf(v).Kind() {
		case reflect.Slice:
			sl := toNativeSlice(normalizeSlice(v))
			res[k] = sl
		case mapType.Kind():
			m, _ := v.(*Map)
			res[k] = ToNativeMap(m)
		default:
			res[k] = v
		}
	}

	return res
}
