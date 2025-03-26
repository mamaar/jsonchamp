package jsonchamp

import (
	"testing"
)

func TestValueGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		setup func() value

		key uint64

		expectedValue any
		expectedOk    bool
	}{
		{
			setup: func() value {
				return value{key: newKey("key", 1), value: "hello"}
			},
			key:           1,
			expectedOk:    true,
			expectedValue: "hello",
		},
		{
			setup: func() value {
				return value{key: newKey("key", 1), value: "world"}
			},
			key:           2,
			expectedOk:    false,
			expectedValue: nil,
		},
	}

	for _, tt := range tests {
		v := tt.setup()
		got, ok := v.get(newKey("key", tt.key))

		if ok != tt.expectedOk {
			t.Errorf("get(%v) = %v; want %v", tt.key, ok, tt.expectedOk)
		}

		if tt.expectedOk && got != tt.expectedValue {
			t.Errorf("get(%v) = %v; want %v", tt.key, got, tt.expectedValue)
		}
	}
}

func TestValueSet(t *testing.T) {
	t.Parallel()

	var v node = value{key: newKey("key", 1<<63), value: "hello"}
	v = v.set(newKey("key", 1<<63), "world")

	world, ok := v.get(newKey("key", 1<<63))
	if !ok {
		t.Errorf("get(1) = %v; want %v", ok, true)
	}

	if world != "world" {
		t.Errorf("get(1) = %v; want %v", world, "world")
	}
}

func TestValueSetWithCollision(t *testing.T) {
	t.Parallel()

	v := value{
		key:   newKey("key", 1),
		value: nil,
	}

	n := v.set(newKey("something", 1), "hello")

	_, isCollision := n.(*collision)
	if !isCollision {
		t.Errorf("set(1) = %v; want %v", isCollision, true)
	}
}
