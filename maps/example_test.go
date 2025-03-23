package maps

import (
	"encoding/json"
	"fmt"
)

func ExampleUnmarshal() {
	const rawJson = `{"name": "John", "age": 30}`

	var rawMap *Map
	if err := json.Unmarshal([]byte(rawJson), &rawMap); err != nil {
		panic(err)
	}

	name, _ := rawMap.Get("name")
	fmt.Println(name)
	// Output: John
}

func ExampleTo() {
	m := New()
	m = m.Set("name", "John")
	m = m.Set("age", 30)

	type Person struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	p, err := To[Person](m)
	if err != nil {
		panic(err)
	}

	fmt.Println(p.Name)
	// Output: John
}
