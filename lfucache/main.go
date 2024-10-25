package main

import (
	"fmt"
	"lfucache/alg"
)

func main() {

	obj := alg.NewLFUCache(2)
	obj.Put(1, 1)
	obj.Put(2, 2)
	res := obj.Get(1)
	fmt.Println(res)
	obj.Put(3, 3)
	res = obj.Get(2)
	fmt.Println(res)
	res = obj.Get(3)
	fmt.Println(res)
	obj.Put(4, 4)
	res = obj.Get(1)
	fmt.Println(res)
	res = obj.Get(3)
	fmt.Println(res)
	res = obj.Get(4)
	fmt.Println(res)
}
