package alg

func MaxProduct(words []string) (int, string) {
	n := len(words)
	if n <= 0 {
		return 0, ""
	}

	masks := make([]int, n)
	lens := make([]int, n)
	stringOut := ""
	for i := 0; i < n; i++ {
		bitmask := 0
		for _, ch := range words[i] {
			bitmask |= 1 << (ch - 'a')
		}
		masks[i] = bitmask
		lens[i] = len(words[i])
	}

	maxVal := 0
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			if masks[i]&masks[j] == 0 {
				stringOut = words[i] + "_" + words[j]
				maxVal = max(maxVal, lens[i]*lens[j])
			}
		}
	}

	return maxVal, stringOut

}

func max(a, b int) int {
	if a > b {
		return a
	}

	return b
}
