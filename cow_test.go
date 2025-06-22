package jsonchamp

import "testing"

func TestSet(t *testing.T) {
	s := newCowSlice()

	if s.Len() != 0 {
		t.Fatal("expects length to be 0")
	}

	s = s.Insert(0, &value{})
	if s.Len() != 1 {
		t.Fatal("expects length to be 1")
	}
}
