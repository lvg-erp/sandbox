package main

import (
	"bullscows/alg"
	"fmt"
)

func main() {
	secret := "1807"
	guess := "2471"

	result := alg.GetHint(secret, guess)
	fmt.Println(result)
}
