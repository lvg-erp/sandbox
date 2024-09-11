package main

import (
	"fmt"
	"wordsearch/alg"
)

func main() {
	board := [][]byte{
		{'o', 'a', 'a', 'n'},
		{'e', 't', 'a', 'e'},
		{'i', 'h', 'k', 'r'},
		{'i', 'f', 'l', 'v'},
	}

	words := []string{"oath", "pea", "eat", "rain"}

	result := alg.FindWords(board, words)

	fmt.Println(result)
}
