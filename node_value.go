package jsonchamp

type node interface {
	get(key key) (any, bool)
	set(key key, value any) node
	copy() node
}

type value struct {
	key   key
	value any
}

// Set implements node.
func (v value) set(key key, newValue any) node {
	// Hash collision for different keys
	if key.hash == v.key.hash && key.key != v.key.key {
		var c node
		c = &collision{
			values: []value{},
		}
		c = c.set(v.key, v.value)
		c = c.set(key, newValue)

		return c
	}

	if key.key != v.key.key {
		panic("key mismatch")
	}

	return value{
		key:   key,
		value: newValue,
	}
}

func (v value) copy() node {
	return value{
		key:   v.key,
		value: v.value,
	}
}

// Get implements node.
func (v value) get(key key) (any, bool) {
	if key == v.key {
		return v.value, true
	}

	return nil, false
}

var _ node = value{
	key: key{
		key:  "",
		hash: 0,
	},
	value: nil,
}
