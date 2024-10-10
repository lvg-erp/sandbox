package alg

import "unicode"

func DecodeString(s string) string {
	stack := []rune{}

	for _, char := range s {
		if char == ']' {
			decodeString := ""
			for stack[len(stack)-1] != '[' {
				decodeString = string(stack[len(stack)-1]) + decodeString
				stack = stack[:len(stack)-1]
			}
			stack = stack[:len(stack)-1]
			base := 1
			k := 0
			for len(stack) > 0 && unicode.IsDigit(stack[len(stack)-1]) {
				k = k + int(stack[len(stack)-1]-'0')*base
				stack = stack[:len(stack)-1]
				base *= 10
			}
			for k > 0 {
				for j := 0; j < len(decodeString); j++ {
					stack = append(stack, rune(decodeString[j]))
				}
				k--
			}
		} else {

			stack = append(stack, char)

		}
	}
	return string(stack)
}
