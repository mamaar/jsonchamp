package maps

import (
	"fmt"
	"reflect"
)

var (
	mapType = reflect.TypeOf(&Map{})
)

func normalizeSlice(in any) []any {
	if reflect.TypeOf(in).Kind() != reflect.Slice {
		return []any{}
	}

	elem := reflect.TypeOf(in).Elem()

	if elem.Kind() == reflect.Interface {
		return in.([]any)
	}

	var result []any
	for i := 0; i < reflect.ValueOf(in).Len(); i++ {
		result = append(result, reflect.ValueOf(in).Index(i).Interface())
	}

	return result
}

func normalizeValue(in any) any {
	t := reflect.TypeOf(in)
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return reflect.ValueOf(in).Int()
	case reflect.Slice:
		return normalizeSlice(in)
	case reflect.String:
		return in.(string)
	case mapType.Kind():
		return in.(*Map)
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
			slicesAreEqual := EqualsAnyList(oneSlice, otherSlice)
			if !slicesAreEqual {
				diff = diff.Set(k, otherSlice)
			}
		default:
			panic(fmt.Sprintf("Unhandled type %T", oneValue))
		}
	}

	return diff
}
