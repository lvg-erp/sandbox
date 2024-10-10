package main

import (
	"decodestring/alg"
	"fmt"
)

func main() {
	s := "3[a]2[bc]"

	res := alg.DecodeString(s)
	fmt.Println(res)
}
