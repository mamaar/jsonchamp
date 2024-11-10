package maps

import (
	"fmt"
	"maps"
	"reflect"
	"slices"

	"gonum.org/v1/gonum/floats/scalar"
)

type ComparableMap[T any] interface {
	Equals(other T) bool
}

func EqualsAnyList(a []any, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 && len(b) == 0 {
		return true
	}
	for i := range a {
		if !EqualsAny(a[i], b[i]) {
			return false
		}
	}
	return true
}

func EqualsStringList(a []string, b []string) bool {
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

func EqualsAny(a any, b any) bool {
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	switch v := a.(type) {
	case string:
		stringEquals := a.(string) == b.(string)
		return stringEquals
	case int:
		intEquals := a.(int) == b.(int)
		return intEquals
	case int32:
		int32Equals := a.(int32) == b.(int32)
		return int32Equals
	case int64:
		int64Equals := a.(int64) == b.(int64)
		return int64Equals
	case float64:
		floatEquals := scalar.EqualWithinAbs(a.(float64), b.(float64), 0.0001)
		return floatEquals
	case *Map:
		bTyped := b.(*Map)
		return v.Equals(bTyped)
	case []any:
		return EqualsAnyList(a.([]any), b.([]any))
	case []string:
		return EqualsStringList(a.([]string), b.([]string))
	case nil:
		return a == nil && b == nil
	default:
		panic(fmt.Sprintf("type %T not supported", a))
	}
}

func Union(one []string, other []string) []string {
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
