package alg

func backtracking(factors []int, ans *[][]int) {
	if len(factors) > 1 {
		combo := make([]int, len(factors))
		copy(combo, factors)
		*ans = append(*ans, combo)
	}

	lastFactor := factors[len(factors)-1]
	factors = factors[:len(factors)-1]
	start := 2
	if len(factors) > 0 {
		start = factors[len(factors)-1]
	}
	for i := start; i*i <= lastFactor; i++ {
		if lastFactor%i == 0 {
			factors = append(factors, i)
			factors = append(factors, lastFactor/i)
			backtracking(factors, ans)
			factors = factors[:len(factors)-2]
		}
	}

	factors = append(factors, lastFactor)

}

func GetFactors(n int) [][]int {
	var ans [][]int
	backtracking([]int{n}, &ans)
	return ans
}
