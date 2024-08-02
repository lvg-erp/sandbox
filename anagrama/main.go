package main

import (
	"anagrama/alg"
	"github.com/labstack/gommon/log"
)

func main() {
	str1 := "banana"
	str2 := "ananab"

	res := alg.IsAnagram_(str1, str2)
	log.Print(res)
}
