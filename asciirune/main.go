package main

import "fmt"

func main() {

	s := []rune("привете")
	for i := 0; i < len(s); i++ {
		fmt.Println(s[i], string(s[i]))
	}
}
