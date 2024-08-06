package alg

import "fmt"

func MaxProfit(prices []int) int {
	i := 0
	valley := prices[0]
	peack := prices[0]
	maxProfit := 0

	for i < len(prices)-1 {
		for i < len(prices)-1 && prices[i] >= prices[i+1] {
			i++
		}
		valley = prices[i]
		fmt.Printf("%s%d", "valley ", valley)
		fmt.Println()
		for i < len(prices)-1 && prices[i] <= prices[i+1] {
			i++
		}
		peack = prices[i]
		fmt.Printf("%s%d", "peack ", peack)
		fmt.Println()
		maxProfit += peack - valley
	}

	return maxProfit
}

// ////variant two//////////////////////////////////////////////////////////////////////
func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func MaxProfit2(prices []int) int {
	length := len(prices)
	if length <= 1 {
		return 0
	}
	leftMin := prices[0]
	rightMax := prices[length-1]
	leftProfits := make([]int, length)
	rightProfits := make([]int, length+1)
	for l := 1; l < length; l++ {
		leftProfits[l] = max(leftProfits[l-1], prices[l]-leftMin)
		leftMin = min(leftMin, prices[l])
		r := length - 1 - l
		rightProfits[r] = max(rightProfits[r+1], rightMax-prices[r])
		rightMax = max(rightMax, prices[r])
	}
	maxProfit := 0
	for i := 0; i < length; i++ {
		maxProfit = max(maxProfit, leftProfits[i]+rightProfits[i+1])
	}

	return maxProfit
}
