package main

import (
	"fmt"
	"stackworktwo/alg"
)

func main() {
	ms := alg.NewStack()
	ms.Push(1)
	ms.Push(2)
	res := ms.Top()
	fmt.Println(res)
	res = ms.Pop()
	fmt.Println(res)
	b := ms.Empty()
	fmt.Println(b)
}
