package main

import (
	"fmt"
)

type DB struct {
	store map[interface{}]interface{}
}

func (d *DB) Put(key interface{}, value interface{}) bool {
	d.store[key] = value
	return true
}

func (d *DB) Get(key interface{}) (interface{}, bool) {
	val, ok := d.store[key]
	if ok {
		return val, true
	} else {
		return val, false
	}
}

func (d *DB) Delete(key interface{}) bool {
	_, ok := d.store[key]
	if ok {
		delete(d.store, key)
		return true
	} else {
		return false
	}
}

func main() {
	sl := []int{1, 2, 3}
	kv_store := DB{store: map[interface{}]interface{}{}}
	kv_store.Put(1, "one")
	kv_store.Put("two", 2)
	kv_store.Put("list", sl)

	key := "list"
	val, ok := kv_store.Get(key)
	if ok {
		fmt.Printf("The value of key %v is %v\n", key, val)
	}

	fmt.Printf("%v\n", kv_store.Delete(1))
	fmt.Printf("%v\n", kv_store.Delete(1))
	fmt.Printf("%v\n", kv_store.Delete(1))

}
