package main

import "fmt"

func main() {
	s := []int{1, 2, 3}
	//s = append(s[:1], s[2:]...)
	s = append(s[:1], append(s[2:], s[1])...)
	fmt.Println(s)
}
