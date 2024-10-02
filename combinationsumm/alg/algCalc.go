package alg

func backtrack(remain int, k int, comb []int, nextStart int, result *[][]int) {
	if remain == 0 && len(comb) == k {
		temp := make([]int, len(comb))
		copy(temp, comb)
		*result = append(*result, temp)
		return
	} else if remain < 0 || len(comb) == k {
		return
	}

	for i := nextStart; i < 9; i++ {
		comb = append(comb, i+1)
		backtrack(remain-i-1, k, comb, i+1, result)
		comb = comb[:len(comb)-1]
	}
}

func CombinationSum(k, n int) [][]int {
	var result [][]int
	var comb []int

	backtrack(n, k, comb, 0, &result)
	return result
}
