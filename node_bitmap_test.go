package jsonchamp

import (
	"fmt"
	"testing"
)

func TestBitmapGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		setuo func() *bitmasked

		key key

		expectedValue any
		expectedOk    bool
	}{
		{
			name: "empty",
			setuo: func() *bitmasked {
				return &bitmasked{
					level:      0,
					valueMap:   0b0000_0000_0000_0000,
					subMapsMap: 0,
					values:     newCowSliceWithItems(&value{key: newKey("key", 1), value: "hello"}),
				}
			},
			key:           newKey("key", 1<<10),
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name: "not found",
			setuo: func() *bitmasked {
				return &bitmasked{
					level:      0,
					valueMap:   0b0000_0001,
					subMapsMap: 0,
					values:     newCowSliceWithItems(&value{key: newKey("key", 1), value: "world"}),
				}
			},
			key:           newKey("key", 2),
			expectedOk:    false,
			expectedValue: nil,
		},
		{
			name: "value found",
			setuo: func() *bitmasked {
				return &bitmasked{
					level:      0,
					valueMap:   0b00000000_00000000_00000000_00000001_00000000_00000000_00000000_00000000,
					subMapsMap: 0,
					values:     newCowSliceWithItems(&value{key: newKey("key", 1<<63), value: "hello"}),
				}
			},
			key:           newKey("key", 1<<63),
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			name: "collision on one section",
			setuo: func() *bitmasked {
				return &bitmasked{
					level:      0,
					valueMap:   0b00000000_00000000_00000000_00000001_00000000_00000000_00000000_00000000,
					subMapsMap: 0,
					values: newCowSliceWithItems(&collision{values: map[string]*value{
						"key_1": {key: newKey("key_1", 1<<63), value: "hello"},
						"key_2": {key: newKey("key_2", 1<<63), value: "world"},
					}}),
				}
			},
			key:           key{key: "key_1", hash: 1 << 63},
			expectedOk:    true,
			expectedValue: "hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			b := tt.setuo()
			got, ok := b.get(tt.key)

			if ok != tt.expectedOk {
				t.Errorf("get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
			}

			if tt.expectedOk && got != tt.expectedValue {
				t.Errorf("get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
			}
		})
	}
}

func TestBitmapSet(t *testing.T) {
	t.Parallel()

	var c node
	c = &bitmasked{
		level:      0,
		valueMap:   0,
		subMapsMap: 0,
		values:     newCowSlice(),
	}

	c = c.set(newKey("key_1", 1<<63), "hello")
	if c.(*bitmasked).values.Len() != 1 {
		t.Fatalf("set(1) = %v; want %v", c.(*bitmasked).values.Len(), 1)
	}

	if _, ok := c.get(newKey("key_1", 1<<63)); !ok {
		t.Fatalf("get(1) = %v; want %v", ok, true)
	}
}

func TestBitmapSetWithCollision(t *testing.T) {
	t.Parallel()

	var c node
	c = &bitmasked{
		level:      0,
		valueMap:   0,
		subMapsMap: 0,
		values:     newCowSlice(),
	}

	c = c.set(newKey("key_1", 1<<63), "hello")
	c = c.set(newKey("key_2", 1<<63|1<<47), "world")

	if c.(*bitmasked).values.Len() != 1 {
		t.Fatalf("set(1) = %v; want %v", c.(*bitmasked).values.Len(), 1)
	}

	if _, ok := c.get(newKey("key_1", 1<<63)); !ok {
		t.Fatalf("get(1) = %v; want %v", ok, true)
	}

	if _, ok := c.get(newKey("key_2", 1<<63|1<<47)); !ok {
		t.Fatalf("get(2) = %v; want %v", ok, true)
	}
}

func TestPartition(t *testing.T) {
	t.Parallel()

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
		t.Run(fmt.Sprintf("hash %d", tt.hash), func(t *testing.T) {
			t.Parallel()

			got := partition(tt.hash, tt.level)
			if got != tt.want {
				t.Errorf("partition(%064b, %d) = %064b; want %064b", tt.hash, tt.level, got, tt.want)
			}
		})
	}
}
func TestPartitionMask(t *testing.T) {
	t.Parallel()

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
		t.Run(fmt.Sprintf("level %d", tt.level), func(t *testing.T) {
			t.Parallel()

			got := partitionMask(tt.level)
			if got != tt.want {
				t.Errorf("partitionMask(%d) = %064b; want %064b", tt.level, got, tt.want)
			}
		})
	}
}

func TestBitmapCopy(t *testing.T) {
	t.Parallel()

	b := &bitmasked{
		level:      0,
		valueMap:   0b00000000_00000000_00000000_00000001_00000000_00000000_00000000_00000000,
		subMapsMap: 0b00000000_00000000_00000000_00000000_00000000_00000000_00000000_00000000,
		values: newCowSliceWithItems(
			&bitmasked{
				level:      1,
				valueMap:   0,
				subMapsMap: 0,
				values:     nil,
			}, &bitmasked{
				level:      2,
				valueMap:   0,
				subMapsMap: 0,
				values:     nil,
			},
		),
	}

	newB := b.copy().(*bitmasked)

	if b.level != newB.level {
		t.Fatalf("level = %d; want %d", newB.level, b.level)
	}

	if b.valueMap != newB.valueMap {
		t.Fatalf("valueMap = %064b; want %064b", newB.valueMap, b.valueMap)
	}

	if b.subMapsMap != newB.subMapsMap {
		t.Fatalf("subMapsMap = %064b; want %064b", newB.subMapsMap, b.subMapsMap)
	}

	if b.values.Get(0).(*bitmasked).level != newB.values.Get(0).(*bitmasked).level {
		t.Fatalf("values[0].level = %d; want %d", newB.values.Get(0).(*bitmasked).level, b.values.Get(0).(*bitmasked).level)
	}

	if b.values.Get(1).(*bitmasked).level != newB.values.Get(1).(*bitmasked).level {
		t.Fatalf("values[1].level = %d; want %d", newB.values.Get(1).(*bitmasked).level, b.values.Get(1).(*bitmasked).level)
	}
}
