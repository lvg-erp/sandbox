package alg

func RemoveDuplicateLetters(s string) string {
	stack := []rune{}
	seen := make(map[rune]bool)
	lastOccurrence := make(map[rune]int)

	for i, c := range s {
		lastOccurrence[c] = i
	}

	for i, c := range s {
		if !seen[c] {
			for len(stack) > 0 && c < stack[len(stack)-1] && i < lastOccurrence[stack[len(stack)-1]] {
				seen[stack[len(stack)-1]] = false
				stack = stack[:len(stack)-1]
			}
			seen[c] = true
			stack = append(stack, c)
		}
	}

	return string(stack)
}
