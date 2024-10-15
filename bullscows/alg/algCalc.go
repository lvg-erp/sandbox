package alg

import "strconv"

func GetHint(secret, guess string) string {

	h := make(map[rune]int)

	for _, ch := range secret {
		h[ch]++
	}

	bulls, cows := 0, 0

	n := len(guess)
	secretArray := []rune(secret)
	guessArray := []rune(guess)

	for idx := 0; idx < n; idx++ {
		ch := guessArray[idx]
		if count, ok := h[ch]; ok {
			if ch == secretArray[idx] {
				bulls++
				if count <= 0 {
					cows--
				}
			} else {
				if count > 0 {
					cows++
				}
			}
			h[ch]--
		}
	}

	return strconv.Itoa(bulls) + "A" + strconv.Itoa(cows) + "B"

}
