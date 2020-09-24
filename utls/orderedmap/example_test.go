/**
 * Copyright 2020 gd Author. All rights reserved.
 * Author: Chuck1024
 */

package orderedmap

import (
	"fmt"
	"testing"
)

func TestGetAndSetExample(t *testing.T) {
	// Init new OrderedMap
	om := NewOrderedMap()

	// Set key
	om.Set("a", 1)
	om.Set("b", 2)
	om.Set("c", 3)
	om.Set("d", 4)

	// Same interface as builtin map
	if val, ok := om.Get("b"); ok == true {
		// Found key "b"
		fmt.Println(val)
	}

	// Delete a key
	om.Delete("c")

	// Failed Get lookup becase we deleted "c"
	if _, ok := om.Get("c"); ok == false {
		// Did not find key "c"
		fmt.Println("c not found")
	}
}

func TestIteratorExample(t *testing.T) {
	n := 100
	om := NewOrderedMap()

	for i := 0; i < n; i++ {
		// Insert data into OrderedMap
		om.Set(i, fmt.Sprintf("%d", i*i))
	}

	// Iterate though values
	// - Values iteration are in insert order
	// - Returned in a key/value pair struct
	iter := om.IterFunc()
	for kv, ok := iter(); ok; kv, ok = iter() {
		fmt.Println(kv, kv.Key, kv.Value)
	}
}

func TestCustomStruct(t *testing.T) {
	om := NewOrderedMap()
	om.Set("one", &MyStruct{1, true})
	om.Set("two", &MyStruct{2, true})
	om.Set("three", &MyStruct{3, true})

	fmt.Println(om)
	// Ouput: OrderedMap[one:&{1 true},  two:&{2 true},  three:&{3 true}, ]
}
