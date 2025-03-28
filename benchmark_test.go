package jsonchamp

import (
	"crypto/rand"
	"fmt"
	"hash"
	"hash/fnv"
	"hash/maphash"
	"strconv"
	"strings"
	"testing"
)

var benchTable = []struct {
	name string
	key  string
}{
	{"short", strings.Repeat("a", 8)},
	{"medium", strings.Repeat("a", 16)},
	{"long", strings.Repeat("a", 24)},
	{"huge", strings.Repeat("a", 128)},
	{"massive", strings.Repeat("a", 1024)},
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

				for i := range b.N {
					m = m.Set(fmt.Sprintf("%s%d", keyLenTT.key, i), keyLenTT.key)
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

				for i := range b.N {
					m = m.Set(fmt.Sprintf("%s%d", keyLenTT.key, i), keyLenTT.key)
				}

				b.ResetTimer()

				for i := range b.N {
					iStr := strconv.Itoa(i)
					k := keyLenTT.key + iStr

					v, ok := m.Get(k)
					if !ok {
						panic(fmt.Sprintf("expected key %s to be in map", k))
					}

					if v != keyLenTT.key {
						panic(fmt.Sprintf("expected %s, got %s", keyLenTT.key, v))
					}
				}
			})
		}
	}
}

func TestBench(t *testing.T) {
	t.Parallel()

	m := New()

	for range 1000_000 {
		k := rand.Text()
		m = m.Set(k, k)

		_, ok := m.Get(k)
		if !ok {
			t.Fatalf("expected key %s to be in map", k)
		}
	}

	t.Logf("map size: %d", len(m.Keys()))
}
