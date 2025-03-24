package maps

type collision struct {
	values []*value
}

func (c *collision) copy() node {
	newValues := make([]*value, len(c.values))
	for i, v := range c.values {
		newValues[i] = &value{
			key:   v.key,
			value: v.value,
		}
	}
	return &collision{values: newValues}
}

// Get implements node.
func (c *collision) get(key Key) (any, bool) {
	for _, v := range c.values {
		if v.key == key {
			return v.value, true
		}
	}
	return nil, false
}

// Set implements node.
func (c *collision) set(key Key, newValue any) node {
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
