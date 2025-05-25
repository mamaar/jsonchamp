package jsonchamp

import (
	"maps"
)

type collision struct {
	values map[string]*value
}

func (c *collision) copy() node {
	newValues := maps.Clone(c.values)

	return &collision{values: newValues}
}

// Get implements node.
func (c *collision) get(key key) (any, bool) {
	for _, v := range c.values {
		if v.key == key {
			return v.value, true
		}
	}

	return nil, false
}

// Set implements node.
func (c *collision) set(key key, newValue any) node {
	newCollision := c.copy().(*collision)
	
	newCollision.values[key.key] = &value{
		key:   key,
		value: newValue,
	}

	return newCollision
}

var _ node = &collision{
	values: nil,
}
