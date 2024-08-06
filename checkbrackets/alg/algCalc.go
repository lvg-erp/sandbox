package alg

import "fmt"

func IsValid(s string) bool {
	pairs := map[string]string{
		")": "(",
		"]": "[",
		"}": "{",
	}

	stack := make([]string, 0)

	if len(s) == 1 {
		return false
	}

	for _, v := range s {
		if string(v) == "(" || string(v) == "[" || string(v) == "{" {
			stack = append(stack, string(v))
		} else {
			if len(stack) == 0 {

				return false

			}
			if pair, ok := pairs[string(v)]; ok && pair != stack[len(stack)-1] {
				fmt.Println(pair)
				return false

			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) != 0 {

		return false

	}

	return true

}
