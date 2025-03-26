package jsonchamp

import (
	"testing"
)

func TestCollisionGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setuo func() *collision

		key uint64

		expectedValue any
		expectedOk    bool
	}{
		{
			setuo: func() *collision {
				return &collision{values: []value{{key: newKey("key", 1), value: "hello"}}}
			},
			key:           1,
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			setuo: func() *collision {
				return &collision{values: []value{{key: newKey("key", 1), value: "world"}}}
			},
			key:           2,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			setuo: func() *collision {
				return &collision{values: []value{
					{key: newKey("key", 1), value: "hello"},
					{key: newKey("key", 2), value: "world"},
				}}
			},
			key:           2,
			expectedOk:    true,
			expectedValue: "world",
		},
	}

	for _, tt := range tests {
		c := tt.setuo()
		got, ok := c.get(newKey("key", tt.key))

		if ok != tt.expectedOk {
			t.Errorf("get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
		}

		if tt.expectedOk && got != tt.expectedValue {
			t.Errorf("get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
		}
	}
}

func TestCollisionSet(t *testing.T) {
	t.Parallel()

	const (
		world = "world"
		hello = "hello"
		key1  = "key_1"
		key2  = "key_2"
	)

	var c node
	c = &collision{values: []value{{key: newKey(key1, 1), value: hello}}}

	c = c.set(newKey(key2, 1), world)

	if len(c.(*collision).values) != 2 {
		t.Errorf("set(2) = %v; want %v", len(c.(*collision).values), 2)
	}

	c = c.set(newKey(key1, 1), world)
	if len(c.(*collision).values) != 2 {
		t.Errorf("set(1) = %v; want %v", len(c.(*collision).values), 2)
	}

	v, ok := c.get(newKey(key1, 1))
	if !ok {
		t.Errorf("get(1) = %v; want %v", ok, true)
	}

	if v != world {
		t.Errorf("get(1) = %v; want %v", v, world)
	}

	v2, ok := c.get(newKey(key2, 1))
	if !ok {
		t.Errorf("get(2) = %v; want %v", ok, true)
	}

	if v2 != world {
		t.Errorf("get(2) = %v; want %v", v2, world)
	}
}
