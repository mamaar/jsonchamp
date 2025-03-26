package jsonchamp

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"gonum.org/v1/gonum/floats/scalar"
)

const (
	floatCompareTolerance = 0.0001
)

func equalsAnyList(a []any, b []any) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	for i := range a {
		if !equalsAny(a[i], b[i]) {
			return false
		}
	}

	return true
}

func equalsStringList(a []string, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	if len(a) == 0 && len(b) == 0 {
		return true
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func toLargestType(a any) any {
	switch v := a.(type) {
	case int:
		return int64(v)
	case int8:
		return int64(v)
	case int16:
		return int64(v)
	case int32:
		return int64(v)
	case int64:
		return v
	case uint:
		return uint64(v)
	case uint8:
		return uint64(v)
	case uint16:
		return uint64(v)
	case uint32:
		return uint64(v)
	case uint64:
		return v
	case float32:
		return float64(v)
	case float64:
		return v
	default:
		return a
	}
}

func equalsAny(a any, b any) bool {
	a = toLargestType(a)
	b = toLargestType(b)

	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}

	switch v := a.(type) {
	case string:
		aString, ok := a.(string)
		if !ok {
			return false
		}

		bString, ok := b.(string)
		if !ok {
			return false
		}

		stringEquals := aString == bString

		return stringEquals
	case int64:
		aInt64, ok := a.(int64)
		if !ok {
			return false
		}

		bInt64, ok := b.(int64)
		if !ok {
			return false
		}

		int64Equals := aInt64 == bInt64

		return int64Equals
	case float64:
		aFloat, ok := a.(float64)
		if !ok {
			return false
		}

		bFloat, ok := b.(float64)
		if !ok {
			return false
		}

		floatEquals := scalar.EqualWithinAbs(aFloat, bFloat, floatCompareTolerance)

		return floatEquals
	case *Map:
		bTyped, ok := b.(*Map)
		if !ok {
			return false
		}

		return v.Equals(bTyped)
	case []any:
		aSlice, ok := a.([]any)
		if !ok {
			return false
		}

		bSlice, ok := b.([]any)
		if !ok {
			return false
		}

		return equalsAnyList(aSlice, bSlice)
	case []string:
		aSlice, ok := a.([]string)
		if !ok {
			return false
		}

		bSlice, ok := b.([]string)
		if !ok {
			return false
		}

		return equalsStringList(aSlice, bSlice)
	case bool:
		aBool, ok := a.(bool)
		if !ok {
			return false
		}

		bBool, ok := b.(bool)
		if !ok {
			return false
		}

		return aBool == bBool
	case nil:
		return a == nil && b == nil
	default:
		panic(fmt.Sprintf("type %T not supported", a))
	}
}

func union(one []string, other []string) []string {
	keyMap := make(map[string]struct{}, len(one))
	for _, k := range one {
		keyMap[k] = struct{}{}
	}

	for _, k := range other {
		keyMap[k] = struct{}{}
	}

	keyIter := maps.Keys(keyMap)

	return slices.Collect(keyIter)
}

func intersection(one []string, other []string) []string {
	var intersections []string

	for _, k := range one {
		for _, j := range other {
			if k == j {
				intersections = append(intersections, k)
			}
		}
	}

	return intersections
}
