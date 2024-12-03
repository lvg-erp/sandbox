package alg

import "math"

func CoinChange(coins []int, amount int) (int, []int) {
	//Массив для отслеживания количества потребуемых монет
	dp := make([]int, amount+1)
	//Массив для номиналов монет
	used := make([]int, amount+1)

	for i := 1; i <= amount; i++ {
		dp[i] = math.MaxInt32 //заполним значением - "невозможно собрать сумму"
	}

	//Для суммы 0 нужно 0 монет
	dp[0] = 0

	for _, coin := range coins {
		//	Обойдем в цикле все монеты
		for j := coin; j <= amount; j++ {
			if dp[j-coin] != math.MaxInt32 {
				dp[j] = min(dp[j], dp[j-coin]+1)
				used[j] = coin
			}
		}
	}
	if dp[amount] == math.MaxInt32 {
		return -1, nil
	}
	result := []int{}

	for amount > 0 {
		result = append(result, used[amount])
		amount -= used[amount]
	}

	return dp[amount], result
}
