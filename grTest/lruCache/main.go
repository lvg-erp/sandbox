package main

import (
	"container/list"
	"fmt"
)

type entry struct {
	key   int
	value int
}

type LRUCache struct {
	capacity int
	cache    map[int]*list.Element // Ключ -> узел списка
	list     *list.List            // Двусвязный список
}

func Constructor(capacity int) LRUCache {
	return LRUCache{
		capacity: capacity,
		cache:    make(map[int]*list.Element),
		list:     list.New(),
	}
}

func (lr *LRUCache) Get(idx int) int {
	if elem, ok := lr.cache[idx]; ok {
		lr.list.MoveToFront(elem)
		return elem.Value.(*entry).value
	}
	return -1
}

func (lr *LRUCache) Put(idx int, value int) {
	if elem, ok := lr.cache[idx]; ok {
		fmt.Printf("Updating key: %d, value: %d\n", idx, value)
		lr.list.MoveToFront(elem)
		elem.Value.(*entry).value = value
		return
	}
	if len(lr.cache) >= lr.capacity {
		last := lr.list.Back()
		if last != nil {
			lastEntry := last.Value.(*entry)
			fmt.Printf("Removing key: %d\n", lastEntry.key)
			delete(lr.cache, lastEntry.key)
			lr.list.Remove(last)
		}
	}
	newEntry := &entry{key: idx, value: value}
	elem := lr.list.PushFront(newEntry)
	lr.cache[idx] = elem
	fmt.Printf("Added key: %d, value: %d\n", idx, value)
}

func printCache(cache LRUCache) {
	fmt.Printf("Cache size: %d\n", len(cache.cache))
	for key, elem := range cache.cache {
		entry := elem.Value.(*entry)
		fmt.Printf("Key: %d, Value: %d\n", key, entry.value)
	}
}

func main() {
	cache := Constructor(5)
	res := cache.Get(1)
	fmt.Println("Get(1):", res)
	printCache(cache)

	cache.Put(1, 1)
	fmt.Println("After Put(1,1):")
	printCache(cache)

	cache.Put(2, 2)
	fmt.Println("After Put(2,2):")
	printCache(cache)

	cache.Put(3, 3)
	fmt.Println("After Put(3,3):")
	printCache(cache)

	cache.Put(4, 4)
	fmt.Println("After Put(4,4):")
	printCache(cache)

	cache.Put(5, 5)
	fmt.Println("After Put(5,5):")
	printCache(cache)

	cache.Put(6, 6)
	fmt.Println("After Put(6,6):")
	printCache(cache)

	res = cache.Get(2)
	fmt.Println("Get(2):", res)
	printCache(cache)
}
