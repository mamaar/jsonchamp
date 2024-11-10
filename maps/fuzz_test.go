package maps

import (
	"fmt"
	"hash"
	"hash/fnv"
	"hash/maphash"
	"strings"
	"testing"
)

func FuzzMaps(f *testing.F) {
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

var benchTable = []struct {
	name      string
	keyLength int
}{
	{"short", 8},
	{"medium", 16},
	{"long", 24},
	{"huge", 128},
	{"massive", 1024},
}

var hasherTable = []struct {
	name   string
	hasher func() hash.Hash64
}{
	{"fnv64", fnv.New64},
	{"maphash", func() hash.Hash64 { return &maphash.Hash{} }},
}

func BenchmarkSet(b *testing.B) {
	for _, keyLenTT := range benchTable {
		for _, hasherTT := range hasherTable {
			b.Run(fmt.Sprintf("%s/%s", hasherTT.name, keyLenTT.name), func(b *testing.B) {
				m := New(WithHasher(hasherTT.hasher))

				key := strings.Repeat("a", keyLenTT.keyLength)

				for i := 0; i < b.N; i++ {
					m = m.Set(fmt.Sprintf("%s%d", key, i), key)
				}
			})
		}
	}
}

func BenchmarkGet(b *testing.B) {
	for _, keyLenTT := range benchTable {
		for _, hasherTT := range hasherTable {
			b.Run(fmt.Sprintf("%s/%s", hasherTT.name, keyLenTT.name), func(b *testing.B) {
				m := New(WithHasher(hasherTT.hasher))

				key := strings.Repeat("a", keyLenTT.keyLength)

				for i := 0; i < b.N; i++ {
					m = m.Set(fmt.Sprintf("%s%d", key, i), key)
				}
				b.ResetTimer()

				for i := 0; i < b.N; i++ {
					v, ok := m.Get(fmt.Sprintf("%s%d", key, i))
					if !ok {
						panic(fmt.Sprintf("expected key %s to be in map", fmt.Sprintf("%s%d", key, i)))
					}

					if v != key {
						panic(fmt.Sprintf("expected %s, got %s", key, v))
					}
				}
			})
		}
	}
}
