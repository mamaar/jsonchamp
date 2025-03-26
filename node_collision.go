package jsonchamp

type collision struct {
	values []value
}

func (c *collision) copy() node {
	newValues := make([]value, len(c.values), len(c.values))
	for i, v := range c.values {
		newValues[i] = value{
			key:   v.key,
			value: v.value,
		}
	}

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
	newCollision := make([]value, len(c.values), len(c.values)+1)

	found := false
	for i, v := range c.values {
		if v.key == key {
			newCollision[i] = value{
				key:   key,
				value: newValue,
			}
			found = true
		} else {
			newCollision[i] = value{
				key:   v.key,
				value: v.value,
			}
		}
	}

	if !found {
		newCollision = append(newCollision, value{
			key:   key,
			value: newValue,
		})
	}

	return &collision{values: newCollision}
}

var _ node = &collision{
	values: nil,
}
