package alg

func HIndex(citation []int) int {
	n := len(citation)

	papers := make([]int, n+1)
	for _, c := range citation {
		if c >= n {
			papers[n]++
		} else {
			papers[c]++
		}
	}

	k := n
	s := papers[n]
	for k > s {
		k--
		s += papers[k]
	}

	return s
}
