package main

import (
	"fmt"
	"stringwork/alg"
)

func main() {
	//1
	testString := "the sky is blue"

	outString := alg.ReversWords(testString)

	fmt.Println(outString)

	//2
	s := "leetcodebeer"
	wordDict := []string{"leet", "code", "break"}

	res := alg.WordBreakAll(s, wordDict)

	fmt.Println(res)
}
