package maps

import (
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"reflect"
	goSlices "slices"
	"strings"
)

var (
	ErrDeleteNotImplemented = fmt.Errorf("delete is not implemented")
	ErrKeyNotFound          = fmt.Errorf("key not found")
	ErrWrongType            = fmt.Errorf("wrong type")
)

type Key struct {
	key  string
	hash uint64
}

func NewKey(key string, hash uint64) Key {
	return Key{key: key, hash: hash}
}

type KeyValue struct {
	Key   Key
	Value any
}

const (
	BitPartitionSize = 6
	PartitionMask    = (1 << BitPartitionSize) - 1
)

type node interface {
	Get(key Key) (any, bool)
	Set(key Key, value any) node
}

type value struct {
	key   Key
	value any
}

// Get implements node.
func (v *value) Get(key Key) (any, bool) {
	if key == v.key {
		return v.value, true
	}
	return nil, false
}

// Set implements node.
func (v *value) Set(key Key, newValue any) node {

	// Hash collision for different keys
	if key.hash == v.key.hash && key.key != v.key.key {
		var c node
		c = &collision{}
		c = c.Set(v.key, v.value)
		c = c.Set(key, newValue)
		return c
	}

	if key.key != v.key.key {
		panic("key mismatch")
	}
	return &value{
		key:   key,
		value: newValue,
	}
}

var _ node = &value{}

type collision struct {
	values []*value
}

// Get implements node.
func (c *collision) Get(key Key) (any, bool) {
	for _, v := range c.values {
		if v.key == key {
			return v.value, true
		}
	}
	return nil, false
}

// Set implements node.
func (c *collision) Set(key Key, newValue any) node {
	newCollision := make([]*value, len(c.values))
	copy(newCollision, c.values)

	for i, v := range c.values {
		if v.key == key {
			newCollision[i] = &value{
				key:   key,
				value: newValue,
			}
			return &collision{values: newCollision}
		}
	}
	newCollision = append(c.values, &value{
		key:   key,
		value: newValue,
	})
	return &collision{values: newCollision}
}

var _ node = &collision{}

type Map struct {
	root   *bitmasked
	hasher hash.Hash64
}

type MapOptions struct {
	hasher func() hash.Hash64
}

var DefaultMapOptions = MapOptions{
	hasher: fnv.New64,
}

type MapOption func(*MapOptions)

func WithHasher(h func() hash.Hash64) MapOption {
	return func(o *MapOptions) {
		o.hasher = h
	}
}

func New(opts ...MapOption) *Map {
	options := DefaultMapOptions
	for _, opt := range opts {
		opt(&options)
	}
	return &Map{
		root: &bitmasked{
			level:      0,
			values:     []node{},
			valueMap:   0,
			subMapsMap: 0,
		},
		hasher: options.hasher(),
	}
}

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

func Copy(in *Map) *Map {
	return in.Copy()
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

func (m *Map) ToMap() map[string]any {
	out := map[string]any{}
	for _, k := range m.Keys() {
		v, _ := m.Get(k)
		switch v := v.(type) {
		case *Map:
			m := v.ToMap()
			out[k] = m
		default:
			out[k] = v
		}
	}
	return out
}

func (m *Map) Copy() *Map {
	newMap := New()
	newMap.hasher = m.hasher
	newMap.root = m.root.copy()
	return newMap
}

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
		if !EqualsAny(fValue, otherValue) {
			return false
		}
	}
	return true
}

func (m *Map) Diff(other *Map) (*Map, error) {
	diff := New()

	oneKeys := m.Keys()
	otherKeys := other.Keys()
	unionKeys := Union(oneKeys, otherKeys)

	if len(unionKeys) == 0 {
		return diff, nil
	}

	for _, k := range unionKeys {
		oneValue, oneExists := m.Get(k)
		otherValue, otherExists := other.Get(k)
		if oneExists && !otherExists {
			diff = diff.Set(k, nil)
		}
		if !oneExists && otherExists {
			diff = diff.Set(k, otherValue)
			continue
		}
		if reflect.TypeOf(oneValue) != reflect.TypeOf(otherValue) {
			diff = diff.Set(k, otherValue)
			continue
		}
		switch oneValue.(type) {
		case *Map:
			oneMap := oneValue.(*Map)
			otherMap := otherValue.(*Map)
			subDiff, err := oneMap.Diff(otherMap)
			if err != nil {
				return nil, err
			}
			if len(subDiff.Keys()) > 0 {
				diff = diff.Set(k, subDiff)
			}
		case string:
			oneString := oneValue.(string)
			otherString := otherValue.(string)
			if oneString != otherString {
				diff = diff.Set(k, otherString)
			}
		case int:
			oneInt := oneValue.(int)
			otherInt := otherValue.(int)
			if oneInt != otherInt {
				diff = diff.Set(k, otherInt)
			}
		}
	}

	return diff, nil
}

func (m *Map) Get(key any) (any, bool) {
	switch k := key.(type) {
	case string:
		return m.root.Get(NewKey(k, m.hash(k)))
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

func (m *Map) GetString(key string) (string, error) {
	v, ok := m.Get(key)
	if !ok {
		return "", fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}
	switch v := v.(type) {
	case string:
		return v, nil
	case int:
		return fmt.Sprint(v), nil
	case float64:
		return fmt.Sprint(v), nil
	default:
		return "", fmt.Errorf("%w: expected string, got %T", ErrWrongType, v)
	}
}

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

func (m *Map) GetInt(key string) (int, error) {
	v, ok := m.Get(key)
	if !ok {
		return 0, fmt.Errorf("%w: '%s'", ErrKeyNotFound, key)
	}
	switch v := v.(type) {
	case int:
		return v, nil
	case float64:
		return int(v), nil
	default:
		return 0, fmt.Errorf("%w: expected int, got %T", ErrWrongType, v)
	}
}

// Set implements node.
func (m *Map) Set(key string, value any) *Map {
	hash := m.hash(key)
	m.root = m.root.Set(NewKey(key, hash), value).(*bitmasked)
	return m
}

func (m *Map) Keys() []string {
	root := m.root
	keys := root.Keys()
	return keys
}

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

func (m *Map) Delete(key string) (*Map, bool) {
	k := NewKey(key, m.hash(key))
	var wasDeleted bool
	m.root, wasDeleted = m.root.Delete(k)

	return m, wasDeleted
}

func (m *Map) Contains(key string) bool {
	_, ok := m.Get(key)
	return ok
}

func (m *Map) Error() string {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Sprintf("error marshalling map: %v", err)
	}
	return string(b)
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

// InformationPaths returns a list of paths to all leaf nodes in a feature
func InformationPaths(f *Map) []string {
	return informationPaths([]string{}, f)
}

func HavePathInCommon(a *Map, b *Map) bool {
	aInformationPaths := InformationPaths(a)
	bInformationPaths := InformationPaths(b)
	hasInCommon := Intersection(aInformationPaths, bInformationPaths)
	return len(hasInCommon) > 0
}

func Intersection(one []string, other []string) []string {
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

func RefToLookup(ref *Map) []string {
	refSource, ok := ref.Get("$ref")
	if !ok {
		return []string{}
	}
	strRefSource, ok := refSource.(string)
	if !ok {
		return []string{}
	}
	parts := strings.Split(strRefSource, "#/")
	if len(parts) != 2 {
		return []string{}
	}
	return strings.Split(parts[1], "/")
}
