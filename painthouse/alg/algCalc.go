package alg

func MinCost(costs [][]int) int {
	n := len(costs)
	dp := make([][3]int, n)
	dp[0] = [3]int{costs[0][0], costs[0][1], costs[0][2]}

	for i := 1; i < n; i++ {
		dp[i][0] = costs[i][0] + min(dp[i-1][1], dp[i-1][2])
		dp[i][1] = costs[i][1] + min(dp[i-1][1], dp[i-1][2])
		dp[i][2] = costs[i][2] + min(dp[i-1][1], dp[i-1][2])
	}

	return minAlt(dp[n-1][0], dp[n-1][1], dp[n-1][2])

}

//func min(a, b int) int{
//	if a<b {
//		return a
//	}
//	return b
//}

func minAlt(a, b, c int) int {
	var m int
	if a < b {
		m = a
	} else {
		m = b
	}
	if m > c {
		m = c
	}
	return m
}
