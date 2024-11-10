package maps

import "reflect"

// diffString checks that the two strings are equal and returns a string containing the other string and the equality result.
func diffString(one string, other string) (string, bool) {
	return other, one != other
}

// diffInt checks that the two integers are equal and returns an int containing the other integer and the equality result.
func diffInt(one int, other int) (int, bool) {
	return other, one != other
}

func mapToAnySlice(in any) []any {
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

// DiffSlice checks that the two slices are equal and returns a slice containing the other slice and whether the slices differ.
func DiffSlice[T comparable](one []T, other []T) ([]T, bool) {
	if len(one) != len(other) {
		return other, true
	}
	for i := range one {
		oneType := reflect.TypeOf(one[i])
		otherType := reflect.TypeOf(other[i])
		if oneType.Kind() != otherType.Kind() {
			return other, true
		}
		switch oneType.Kind() {
		case reflect.Slice:
			oneSlice := mapToAnySlice(one[i])
			otherSlice := mapToAnySlice(other[i])
			_, hasDiff := DiffSlice(oneSlice, otherSlice)
			if hasDiff {
				return other, true
			}
			return other, false
		default:
			if one[i] != other[i] {
				return other, true
			}
		}
	}
	return other, false
}
