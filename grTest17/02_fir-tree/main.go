package main

import "fmt"

func main() {
	height := 8
	for i := 0; i <= height; i++ {
		for s := 0; s < height-i; s++ {
			fmt.Print(" ")
		}

		for j := 0; j < 2*i-1; j++ {
			fmt.Print("*")
		}
		fmt.Println()
	}

	for i := 0; i < 2; i++ {
		for s := 0; s < height-1; s++ {
			fmt.Print(" ")
		}
		fmt.Println("|")
	}
}
