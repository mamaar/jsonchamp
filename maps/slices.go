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

func EqualsAny(a any, b any) bool {
	a = toLargestType(a)
	b = toLargestType(b)
	if reflect.TypeOf(a) != reflect.TypeOf(b) {
		return false
	}
	switch v := a.(type) {
	case string:
		stringEquals := a.(string) == b.(string)
		return stringEquals
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
	case bool:
		return a.(bool) == b.(bool)
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
