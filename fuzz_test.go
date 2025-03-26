package jsonchamp

import (
	"testing"
)

func FuzzMapKey(f *testing.F) {
	m := New()

	f.Fuzz(func(t *testing.T, key string) {
		m = m.Set(key, key)

		v, ok := m.Get(key)
		if !ok {
			t.Fatalf("expected key %s to be in map", key)
		}

		if v != key {
			t.Fatalf("expected %s, got %s", key, v)
		}
	})
}

func FuzzMapInt64(f *testing.F) {
	key := "key"
	m := New()

	f.Fuzz(func(t *testing.T, value int64) {
		m = m.Set(key, value)

		v, err := m.GetInt(key)
		if err != nil {
			t.Fatalf("expected key %s to be in map: %v", key, err)
		}

		if v != value {
			t.Fatalf("expected %d, got %d", value, v)
		}
	})
}

func FuzzMapFloat(f *testing.F) {
	key := "key"
	m := New()

	f.Fuzz(func(t *testing.T, value float64) {
		m = m.Set(key, value)

		v, err := m.GetFloat(key)
		if err != nil {
			t.Fatalf("expected key %s to be in map: %v", key, err)
		}

		if v != value {
			t.Fatalf("expected %f, got %f", value, v)
		}
	})
}
