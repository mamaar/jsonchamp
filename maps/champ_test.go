package maps

import "testing"

func TestValueGet(t *testing.T) {
	tests := []struct {
		setup func() *value

		key uint64

		expectedValue any
		expectedOk    bool
	}{
		{
			setup: func() *value {
				return &value{key: NewKey("key", 1), value: "hello"}
			},
			key:           1,
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			setup: func() *value {
				return &value{key: NewKey("key", 1), value: "world"}
			},
			key:           2,
			expectedOk:    false,
			expectedValue: nil,
		},
	}

	for _, tt := range tests {
		v := tt.setup()
		got, ok := v.Get(NewKey("key", tt.key))

		if ok != tt.expectedOk {
			t.Errorf("Get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
		}

		if tt.expectedOk && got != tt.expectedValue {
			t.Errorf("Get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
		}

	}
}

func TestValueSet(t *testing.T) {

	var v node = &value{key: NewKey("key", 1<<63), value: "hello"}
	v = v.Set(NewKey("key", 1<<63), "world")

	world, ok := v.Get(NewKey("key", 1<<63))
	if !ok {
		t.Errorf("Get(1) = %v; want %v", ok, true)
	}

	if world != "world" {
		t.Errorf("Get(1) = %v; want %v", world, "world")
	}
}

func TestValueSetWithCollision(t *testing.T) {
	v := &value{key: NewKey("key", 1)}

	n := v.Set(NewKey("something", 1), "hello")

	_, isCollision := n.(*collision)
	if !isCollision {
		t.Errorf("Set(1) = %v; want %v", isCollision, true)
	}
}

func TestCollisionGet(t *testing.T) {
	tests := []struct {
		setuo func() *collision

		key uint64

		expectedValue any
		expectedOk    bool
	}{
		{
			setuo: func() *collision {
				return &collision{values: []*value{{key: NewKey("key", 1), value: "hello"}}}
			},
			key:           1,
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			setuo: func() *collision {
				return &collision{values: []*value{{key: NewKey("key", 1), value: "world"}}}
			},
			key:           2,
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			setuo: func() *collision {
				return &collision{values: []*value{{key: NewKey("key", 1), value: "hello"}, {key: NewKey("key", 2), value: "world"}}}
			},
			key:           2,
			expectedOk:    true,
			expectedValue: "world",
		},
	}

	for _, tt := range tests {
		c := tt.setuo()
		got, ok := c.Get(NewKey("key", tt.key))

		if ok != tt.expectedOk {
			t.Errorf("Get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
		}

		if tt.expectedOk && got != tt.expectedValue {
			t.Errorf("Get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
		}

	}
}

func TestCollisionSet(t *testing.T) {
	var c node
	c = &collision{values: []*value{{key: NewKey("key_1", 1), value: "hello"}}}

	c = c.Set(NewKey("key_2", 1), "world")

	if len(c.(*collision).values) != 2 {
		t.Errorf("Set(2) = %v; want %v", len(c.(*collision).values), 2)
	}

	c = c.Set(NewKey("key_1", 1), "world")
	if len(c.(*collision).values) != 2 {
		t.Errorf("Set(1) = %v; want %v", len(c.(*collision).values), 2)
	}

	v, ok := c.Get(NewKey("key_1", 1))
	if !ok {
		t.Errorf("Get(1) = %v; want %v", ok, true)
	}
	if v != "world" {
		t.Errorf("Get(1) = %v; want %v", v, "world")
	}

	v2, ok := c.Get(NewKey("key_2", 1))
	if !ok {
		t.Errorf("Get(2) = %v; want %v", ok, true)
	}
	if v2 != "world" {
		t.Errorf("Get(2) = %v; want %v", v2, "world")
	}
}

func TestBitmapGet(t *testing.T) {
	tests := []struct {
		name string

		setuo func() *bitmasked

		key Key

		expectedValue any
		expectedOk    bool
	}{
		{
			name: "empty",
			setuo: func() *bitmasked {
				return &bitmasked{
					valueMap: 0b0000_0000_0000_0000,
					values:   []node{&value{key: NewKey("key", 1), value: "hello"}},
				}
			},
			key:           NewKey("key", 1<<10),
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name: "not found",
			setuo: func() *bitmasked {
				return &bitmasked{
					valueMap: 0b0000_0001,
					values:   []node{&value{key: NewKey("key", 1), value: "world"}},
				}
			},
			key:           NewKey("key", 2),
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name: "value found",
			setuo: func() *bitmasked {
				return &bitmasked{
					valueMap: 0b00000000_00000000_00000000_00000001_00000000_00000000_00000000_00000000,
					values:   []node{&value{key: NewKey("key", 1<<63), value: "hello"}},
				}
			},
			key:           NewKey("key", 1<<63),
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			name: "collision on one section",
			setuo: func() *bitmasked {
				return &bitmasked{
					valueMap: 0b00000000_00000000_00000000_00000001_00000000_00000000_00000000_00000000,
					values:   []node{&collision{values: []*value{{key: NewKey("key_1", 1<<63), value: "hello"}, {key: NewKey("key_2", 1<<63), value: "world"}}}},
				}
			},
			key:           Key{key: "key_1", hash: 1 << 63},
			expectedOk:    true,
			expectedValue: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := tt.setuo()
			got, ok := b.Get(tt.key)

			if ok != tt.expectedOk {
				t.Errorf("Get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
			}

			if tt.expectedOk && got != tt.expectedValue {
				t.Errorf("Get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
			}
		})

	}
}

func TestBitmapSet(t *testing.T) {
	var c node
	c = &bitmasked{values: []node{}}

	c = c.Set(NewKey("key_1", 1<<63), "hello")
	if len(c.(*bitmasked).values) != 1 {
		t.Fatalf("Set(1) = %v; want %v", len(c.(*bitmasked).values), 1)
	}

	if _, ok := c.Get(NewKey("key_1", 1<<63)); !ok {
		t.Fatalf("Get(1) = %v; want %v", ok, true)
	}
}

func TestBitmapSetWithCollision(t *testing.T) {
	var c node
	c = &bitmasked{values: []node{}}

	c = c.Set(NewKey("key_1", 1<<63), "hello")
	c = c.Set(NewKey("key_2", 1<<63|1<<47), "world")

	if len(c.(*bitmasked).values) != 1 {
		t.Fatalf("Set(1) = %v; want %v", len(c.(*bitmasked).values), 1)
	}

	if _, ok := c.Get(NewKey("key_1", 1<<63)); !ok {
		t.Fatalf("Get(1) = %v; want %v", ok, true)
	}

	if _, ok := c.Get(NewKey("key_2", 1<<63|1<<47)); !ok {
		t.Fatalf("Get(2) = %v; want %v", ok, true)
	}
}

func TestPartition(t *testing.T) {
	tests := []struct {
		hash  uint64
		level uint8
		want  uint64
	}{
		{
			hash:  1 << 63,
			level: 0,
			want:  0b00100000,
		},
		{
			hash:  1 << 57,
			level: 1,
			want:  0b00100000,
		},
		{
			hash:  0b0001_0000,
			level: 10,
			want:  0b0000_0000,
		},
		{
			hash:  4,
			level: 10,
			want:  0b0000_0100,
		},
	}

	for _, tt := range tests {
		got := partition(tt.hash, tt.level)
		if got != tt.want {
			t.Errorf("partition(%064b, %d) = %064b; want %064b", tt.hash, tt.level, got, tt.want)
		}
	}
}
func TestPartitionMask(t *testing.T) {
	tests := []struct {
		level uint8
		want  uint64
	}{
		{
			level: 0,
			want:  0b11111100_00000000_00000000_00000000_00000000_00000000_00000000_00000000,
		},
		{
			level: 1,
			want:  0b00000011_11110000_00000000_00000000_00000000_00000000_00000000_00000000,
		},
		{
			level: 2,
			want:  0b00000000_00001111_11000000_00000000_00000000_00000000_00000000_00000000,
		},
		{
			level: 3,
			want:  0b00000000_00000000_00111111_00000000_00000000_00000000_00000000_00000000,
		},
		{
			level: 4,
			want:  0b00000000_00000000_00000000_11111100_00000000_00000000_00000000_00000000,
		},
		{
			level: 5,
			want:  0b00000000_00000000_00000000_00000011_11110000_00000000_00000000_00000000,
		},
		{
			level: 6,
			want:  0b00000000_00000000_00000000_00000000_00001111_11000000_00000000_00000000,
		},
		{
			level: 7,
			want:  0b00000000_00000000_00000000_00000000_00000000_00111111_00000000_00000000,
		},
		{
			level: 8,
			want:  0b00000000_00000000_00000000_00000000_00000000_00000000_11111100_00000000,
		},
		{
			level: 9,
			want:  0b00000000_00000000_00000000_00000000_00000000_00000000_00000011_11110000,
		},
		{
			level: 10,
			want:  0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_00001111,
		},
	}

	for _, tt := range tests {
		got := partitionMask(tt.level)
		if got != tt.want {
			t.Errorf("partitionMask(%d) = %064b; want %064b", tt.level, got, tt.want)
		}
	}
}

func TestDelete(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		setup       func() *Map
		key         string
		expectedMap *Map
	}{
		{
			name: "delete from empty map",
			setup: func() *Map {
				return New()
			},
			key:         "key",
			expectedMap: New(),
		},
		{
			name: "delete from map with one value",
			setup: func() *Map {
				return NewFromItems("key", 123)
			},
			key:         "key",
			expectedMap: New(),
		},
		{
			name: "delete from map with multiple values",
			setup: func() *Map {
				return NewFromItems("key", 123, "key_2", 456)
			},
			key:         "key",
			expectedMap: NewFromItems("key_2", 456),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			m := testCase.setup()
			m, _ = m.Delete(testCase.key)

			if !m.Equals(testCase.expectedMap) {
				t.Errorf("Delete(%s) = %v; want %v", testCase.key, m, testCase.expectedMap)
			}
		})
	}
}
