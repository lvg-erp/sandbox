package main

import (
	"fmt"
	"moneymaxcount/alg"
)

func main() {
	coins := []int{1, 2, 5}
	amount := 11

	quantity, pars := alg.CoinChange(coins, amount)
	fmt.Println(quantity)
	fmt.Println(pars)
}
