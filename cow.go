package jsonchamp

type cowSlice struct {
	slice  []node
	shared bool
}

func newCowSlice() *cowSlice {
	return &cowSlice{
		slice:  nil,
		shared: false,
	}
}

func newCowSliceWithItems(items ...node) *cowSlice {
	return &cowSlice{
		slice:  items,
		shared: false,
	}
}

func (c *cowSlice) Len() int {
	return len(c.slice)
}

func (c *cowSlice) Values() []node {
	return c.slice
}

func (c *cowSlice) Get(i int) node {
	return c.slice[i]
}

func (c *cowSlice) Set(i int, v node) *cowSlice {
	if c.shared {
		n := make([]node, len(c.slice))
		copy(n, c.slice)
		n[i] = v
		return &cowSlice{
			slice:  n,
			shared: false,
		}
	}

	c.slice[i] = v
	return c
}

func (c *cowSlice) Insert(i int, v node) *cowSlice {
	if c.shared {
		n := make([]node, len(c.slice)+1)
		copy(n[:i], c.slice[:i])
		copy(n[i+1:], c.slice[i:])
		n[i] = v
		return &cowSlice{
			slice:  n,
			shared: false,
		}
	}

	c.slice = append(c.slice[:i], append([]node{v}, c.slice[i:]...)...)
	return c
}

func (c *cowSlice) Delete(i int) *cowSlice {
	if c.shared {
		newData := make([]node, len(c.slice))
		copy(newData[:i], c.slice[:i])
		copy(newData[i:], c.slice[i+1:])
		return &cowSlice{
			slice:  newData,
			shared: false,
		}
	}

	c.slice = append(c.slice[:i], c.slice[i+1:]...)
	return c
}

func (c *cowSlice) Share() *cowSlice {
	c.shared = true
	return c
}
