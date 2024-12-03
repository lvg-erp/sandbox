package alg

func ToHex(num int) string {

	if num == 0 {
		return "0"
	}

	hexChars := "0123456789abcdef"

	if num < 0 {
		num += 1 << 32
	}

	result := []byte{}

	for num > 0 {
		result = append(result, hexChars[num%16])
		num /= 16
	}

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return string(result)
}
