package alg

func getNext(n int) int {
	totalSum := 0
	for n > 0 {
		digit := n % 10
		n /= 10
		totalSum += digit * digit
	}
	return totalSum
}

func IsHappy(n int) bool {
	seen := make(map[int]bool)

	for n != 1 && !seen[n] {
		seen[n] = true
		n = getNext(n)
	}

	return n == 1
}
