package main

import (
	"fmt"
	"stringworktwo/alg"
)

func main() {

	input := "eeetcode"
	//hm := alg.String2Map(input)
	//fmt.Println(hm)
	res := alg.SearchUnicodeCharacter(input)
	fmt.Println(res)

}
