package main

import (
	"besttimeaction/alg"
	"fmt"
)

func main() {

	//prices := []int{7, 1, 5, 3, 6, 4}
	//result := alg.MaxProfit(prices)
	//fmt.Println(result)

	prices := []int{3, 3, 5, 0, 0, 3, 1, 4}
	result := alg.MaxProfit2(prices)
	fmt.Println(result)

}
