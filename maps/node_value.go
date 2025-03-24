package maps

// Set implements node.
func (v *value) set(key Key, newValue any) node {

	// Hash collision for different keys
	if key.hash == v.key.hash && key.key != v.key.key {
		var c node
		c = &collision{}
		c = c.set(v.key, v.value)
		c = c.set(key, newValue)
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

type node interface {
	get(key Key) (any, bool)
	set(key Key, value any) node
	copy() node
}

type value struct {
	key   Key
	value any
}

func (v *value) copy() node {
	return &value{
		key:   v.key,
		value: v.value,
	}
}

// Get implements node.
func (v *value) get(key Key) (any, bool) {
	if key == v.key {
		return v.value, true
	}
	return nil, false
}

var _ node = &value{}
