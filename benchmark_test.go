package jsonchamp

import (
	"fmt"
	"hash"
	"hash/fnv"
	"hash/maphash"
	"strconv"
	"strings"
	"testing"
)

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

				for i := range b.N {
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

				for i := range b.N {
					m = m.Set(fmt.Sprintf("%s%d", key, i), key)
				}

				b.ResetTimer()

				for i := range b.N {
					iStr := strconv.Itoa(i)
					k := key + iStr
					v, ok := m.Get(k)
					if !ok {
						panic(fmt.Sprintf("expected key %s to be in map", k))
					}

					if v != key {
						panic(fmt.Sprintf("expected %s, got %s", key, v))
					}
				}
			})
		}
	}
}

func TestBench(t *testing.T) {
	m := New()
	for i := range 1_000 {
		k := fmt.Sprintf("%d", i)
		m = m.Set(k, i)

		_, ok := m.Get(k)
		if !ok {
			t.Fatalf("expected key %s to be in map", k)
		}
	}

	t.Logf("map size: %d", len(m.Keys()))
}
