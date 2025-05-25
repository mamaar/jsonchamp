package jsonchamp

import (
	"errors"
	"fmt"
	"hash"
	"hash/maphash"
	goSlices "slices"
	"strconv"
	"strings"
)

var (
	// ErrKeyNotFound is returned when a key is not found in a map.
	ErrKeyNotFound = errors.New("key not found")
	// ErrWrongType is returned when the value of a key is not of the expected type.
	ErrWrongType = errors.New("wrong type")
)

type key struct {
	key  string
	hash uint64
}

func newKey(k string, hash uint64) key {
	return key{key: k, hash: hash}
}

// Map is an immutable hash map implementation.
type Map struct {
	root   *bitmasked
	hasher hash.Hash64
}

// mapOptions.
type mapOptions struct {
	hasher func() hash.Hash64
}

// defaultMapOptions are the default options used to create a map.
var defaultMapOptions = mapOptions{
	//hasher: fnv.New64,
	hasher: func() hash.Hash64 { return &maphash.Hash{} },
}

// MapOption is a function that sets an option on a map.
type MapOption func(*mapOptions)

// WithHasher sets the hasher used to hash keys in the map.
func WithHasher(h func() hash.Hash64) MapOption {
	return func(o *mapOptions) {
		o.hasher = h
	}
}

// New creates a new map.
func New(opts ...MapOption) *Map {
	options := defaultMapOptions
	for _, opt := range opts {
		opt(&options)
	}

	return &Map{
		root: &bitmasked{
			level:      0,
			values:     make([]node, 0, 5),
			valueMap:   0,
			subMapsMap: 0,
		},
		hasher: options.hasher(),
	}
}

// NewFromItems creates a new map from a list of key-value pairs.
func NewFromItems(items ...any) *Map {
	newMap := New()

	for i := 0; i < len(items); i += 2 {
		key := items[i].(string)

		var value any
		if i+1 >= len(items) {
			value = nil
		} else {
			value = items[i+1]
		}

		newMap = newMap.Set(key, value)
	}

	return newMap
}

func (m *Map) hash(key string) uint64 {
	_, err := m.hasher.Write([]byte(key))
	if err != nil {
		panic(err)
	}

	sum := m.hasher.Sum64()
	m.hasher.Reset()

	return sum
}

// ToMap returns a native Go map with the same structure as the map.
func (m *Map) ToMap() map[string]any {
	return ToNativeMap(m)
}

// Copy returns a deep copy of a map.
func (m *Map) Copy() *Map {
	newMap := New()
	newMap.hasher = m.hasher

	newRoot, ok := m.root.copy().(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	newMap.root = newRoot

	return newMap
}

// Equals compares two maps recursively and returns true if they are equal.
func (m *Map) Equals(other *Map) bool {
	fKeys := m.Keys()
	otherKeys := other.Keys()

	if len(fKeys) != len(otherKeys) {
		return false
	}

	if len(fKeys) == 0 && len(otherKeys) == 0 {
		return true
	}

	for _, k := range fKeys {
		fValue, _ := m.Get(k)

		otherValue, otherExists := other.Get(k)
		if !otherExists {
			return false
		}

		if !equalsAny(fValue, otherValue) {
			return false
		}
	}

	return true
}

func castPair[T any](a any, b any) (T, T) {
	var zero T

	aT, ok := a.(T)
	if !ok {
		return zero, zero
	}

	bT, ok := b.(T)
	if !ok {
		return zero, zero
	}

	return aT, bT
}

// Diff returns a map with the differences between two maps.
// The returned map will contain the keys that are not equal in the two maps.
// The value of the returned map will be the value of the key of the only map that contains it.
// If the value of a key exists in both maps, the function will compare the values
// and return the other value if they are different.
// If the value of a key is a map in both maps, the function will compare the maps recursively.
func (m *Map) Diff(other *Map) *Map {
	return diffMap(m, other)
}

// Get retrieves the value of a key from a map.
// The key can be a string or a list of strings.
// If the key is a list of strings, the function will traverse the map recursively folloeing the keys in the list.
func (m *Map) Get(key any) (any, bool) {
	switch k := key.(type) {
	case string:
		return m.root.get(newKey(k, m.hash(k)))
	case []string:
		if len(k) == 0 {
			return nil, false
		}

		if len(k) == 1 {
			return m.Get(k[0])
		}

		firstKey := k[0]
		restKeys := k[1:]

		firstValue, ok := m.Get(firstKey)
		if !ok {
			return nil, false
		}

		firstValueMap, ok := firstValue.(*Map)
		if !ok {
			return nil, false
		}

		return firstValueMap.Get(restKeys)
	default:
		return nil, false
	}
}

// GetMap retrieves the value of a key from a map and casts it to a map.
func (m *Map) GetMap(key string) (*Map, error) {
	v, ok := m.Get(key)
	if !ok {
		return nil, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	valueMap, ok := v.(*Map)
	if !ok {
		return nil, ErrWrongType
	}

	return valueMap, nil
}

// GetString retrieves the value of a key from a map and casts it to a string.
func (m *Map) GetString(key string) (string, error) {
	v, ok := m.Get(key)
	if !ok {
		return "", fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	switch v := v.(type) {
	case string:
		return v, nil
	case int:
		return strconv.Itoa(v), nil
	case float64:
		return fmt.Sprint(v), nil
	default:
		return "", fmt.Errorf("%w: expected string, got %T", ErrWrongType, v)
	}
}

// GetBool retrieves the value of a key from a map and casts it to a bool.
func (m *Map) GetBool(key string) (bool, error) {
	v, ok := m.Get(key)
	if !ok {
		return false, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	valueBool, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%w: expected bool, got %T", ErrWrongType, v)
	}

	return valueBool, nil
}

// GetFloat retrieves the value of a key from a map and casts it to a float64.
func (m *Map) GetFloat(key string) (float64, error) {
	v, ok := m.Get(key)
	if !ok {
		return 0, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	valueFloat, ok := v.(float64)
	if !ok {
		return 0, fmt.Errorf("%w: expected int, got %T", ErrWrongType, v)
	}

	return valueFloat, nil
}

// GetInt retrieves the value of a key from a map and casts it to an int.
func (m *Map) GetInt(key string) (int64, error) {
	v, ok := m.Get(key)
	if !ok {
		return 0, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	switch v := v.(type) {
	case int64:
		return v, nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("%w: expected int64, got %T", ErrWrongType, v)
	}
}

// Set implements node.
func (m *Map) Set(key string, value any) *Map {
	value = normalizeValue(value)
	h := m.hash(key)

	newRoot, ok := m.root.set(newKey(key, h), value).(*bitmasked)
	if !ok {
		panic("expected bitmasked")
	}

	m.root = newRoot

	return m
}

// Keys returns a list of all keys in the map.
func (m *Map) Keys() []string {
	root := m.root
	keys := root.keys()

	return keys
}

// Merge merges two maps.
// If a key exists in both maps, the value from the other map will be used.
// If the value of a key is a map in both maps, the maps will be merged recursively.
func (m *Map) Merge(other *Map) *Map {
	newMap := m.Copy()

	for _, k := range other.Keys() {
		v, _ := other.Get(k)
		if _, ok := m.Get(k); !ok {
			newMap = newMap.Set(k, v)

			continue
		}

		switch t := v.(type) {
		case *Map:
			// Try to deep merge if both are maps. If current is  not a map, we will overwrite it.
			currentValue, currentValueExists := newMap.Get(k)
			currentValueTyped, currentValueIsMap := currentValue.(*Map)

			if currentValueIsMap && currentValueExists {
				t = currentValueTyped.Merge(t)
				newMap = newMap.Set(k, t.Copy())

				continue
			}

			newMap = newMap.Set(k, t)
		default:
			newMap = newMap.Set(k, v)
		}
	}

	return newMap
}

// Delete removes a key from the map and returns a new map without the key.
func (m *Map) Delete(key string) (*Map, bool) {
	n := m.Copy()
	k := newKey(key, n.hash(key))

	var wasDeleted bool
	n.root, wasDeleted = n.root.delete(k)

	return n, wasDeleted
}

// Contains returns true if a key exists in the map.
func (m *Map) Contains(key string) bool {
	_, ok := m.Get(key)

	return ok
}

func informationPaths(initial []string, f *Map) []string {
	result := initial
	keys := f.Keys()
	keyIter := goSlices.Values(keys)

	keys = goSlices.Sorted(keyIter)
	for _, k := range keys {
		v, _ := f.Get(k)
		switch v := v.(type) {
		case *Map:
			currentPath := []string{k}
			children := informationPaths(currentPath, v)
			concatenated := strings.Join(children, ".")
			result = append(result, concatenated)
		default:
			result = append(result, k)
		}
	}

	return result
}

// InformationPaths returns a list of paths to all leaf nodes in a map.
// The format of the paths is a dot-separated string of keys.
// For example, the following map:
//
//	{
//	    "a": {
//	        "b": 1,
//	        "c": {
//	            "d": 2
//	        }
//	    }
//	}
//
// will return the following paths:
//
//	["a.b", "a.c.d"]
func InformationPaths(f *Map) []string {
	return informationPaths([]string{}, f)
}

// havePathInCommon returns true if two maps have any paths to leaf nodes in common.
func havePathInCommon(a *Map, b *Map) bool {
	aInformationPaths := InformationPaths(a)
	bInformationPaths := InformationPaths(b)
	hasInCommon := intersection(aInformationPaths, bInformationPaths)

	return len(hasInCommon) > 0
}

// Get retrieves the value of a key from a map and casts it to the desired type.
// It can return errors if the key is not found or if the value is not of the expected type.
func Get[T any](m *Map, key string) (T, error) {
	v, ok := m.Get(key)
	if !ok {
		var zero T

		return zero, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}

	casted, ok := v.(T)
	if !ok {
		var zero T

		return zero, fmt.Errorf("%w: '%s'", ErrWrongType, key)
	}

	return casted, nil
}
