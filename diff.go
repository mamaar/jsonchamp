package jsonchamp

import (
	"fmt"
	"reflect"
)

var (
	mapType = reflect.TypeOf(&Map{
		root:   nil,
		hasher: nil,
	})
)

func normalizeSlice(in any) []any {
	if reflect.TypeOf(in).Kind() != reflect.Slice {
		return []any{}
	}

	result := make([]any, 0, reflect.ValueOf(in).Len())
	for i := range reflect.ValueOf(in).Len() {
		result = append(result, reflect.ValueOf(in).Index(i).Interface())
	}

	return result
}

func normalizeNativeMap(in map[string]any) map[string]any {
	out := make(map[string]any, len(in))

	for k, v := range in {
		switch t := v.(type) {
		case map[string]any:
			out[k] = normalizeNativeMap(t)
		case []map[string]any:
			var arr []any
			for _, m := range t {
				arr = append(arr, normalizeValue(m))
			}

			out[k] = arr
		default:
			out[k] = normalizeValue(t)
		}
	}

	return out
}

func normalizeValue(in any) any {
	if in == nil {
		return nil
	}

	t := reflect.TypeOf(in)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(in).Int()
	case reflect.Float32, reflect.Float64:
		return reflect.ValueOf(in).Float()
	case reflect.Slice:
		return normalizeSlice(in)
	case reflect.String:
		stringed, ok := in.(string)
		if !ok {
			panic("could not cast to string")
		}

		return stringed
	case mapType.Kind():
		casted, ok := in.(*Map)
		if !ok {
			panic("could not cast to *Map")
		}

		return casted
	case reflect.Map:
		mapped, ok := in.(map[string]any)
		if !ok {
			panic("could not cast to map[string]any")
		}

		return normalizeNativeMap(mapped)
	default:
		panic(fmt.Sprintf("unsupported type to normalize %v", t))
	}
}

func diffMap(m *Map, other *Map) *Map {
	diff := New()

	oneKeys := m.Keys()
	otherKeys := other.Keys()
	unionKeys := union(oneKeys, otherKeys)

	if len(unionKeys) == 0 {
		return diff
	}

	for _, k := range unionKeys {
		oneValue, oneExists := m.Get(k)
		otherValue, otherExists := other.Get(k)

		if oneExists && !otherExists {
			diff = diff.Set(k, nil)

			continue
		}

		if !oneExists && otherExists {
			diff = diff.Set(k, otherValue)

			continue
		}

		oneValue = normalizeValue(oneValue)
		otherValue = normalizeValue(otherValue)

		if reflect.TypeOf(oneValue) != reflect.TypeOf(otherValue) {
			diff = diff.Set(k, otherValue)

			continue
		}

		oneValue, otherValue = toLargestType(oneValue), toLargestType(otherValue)
		switch oneValue.(type) {
		case *Map:
			oneMap, otherMap := castPair[*Map](oneValue, otherValue)

			subDiff := diffMap(oneMap, otherMap)
			if len(subDiff.Keys()) > 0 {
				diff = diff.Set(k, subDiff)
			}
		case string:
			oneString, otherString := castPair[string](oneValue, otherValue)
			if oneString != otherString {
				diff = diff.Set(k, otherString)
			}
		case int64:
			oneInt, otherInt := castPair[int64](oneValue, otherValue)
			if oneInt != otherInt {
				diff = diff.Set(k, otherInt)
			}
		case float64:
			oneFloat, otherFloat := castPair[float64](oneValue, otherValue)
			if oneFloat != otherFloat {
				diff = diff.Set(k, otherFloat)
			}
		case []any:
			oneSlice, otherSlice := castPair[[]any](oneValue, otherValue)
			slicesAreEqual := equalsAnyList(oneSlice, otherSlice)

			if !slicesAreEqual {
				diff = diff.Set(k, otherSlice)
			}
		default:
			panic(fmt.Sprintf("Unhandled type %T", oneValue))
		}
	}

	return diff
}
