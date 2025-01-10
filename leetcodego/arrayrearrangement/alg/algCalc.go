package alg

import "sort"

func Zip(arrs [][]int) [][]int {
	lengthArray := make([]int, len(arrs))

	for i := 0; i < len(arrs); i++ {
		lengthArray[i] = len(arrs[i])
	}

	sort.Ints(lengthArray)
	minLength := lengthArray[0]
	result := make([][]int, minLength)
	for i := 0; i < minLength; i++ {
		temp := make([]int, len(arrs))
		for j := 0; j < len(arrs); j++ {
			temp[j] = arrs[j][i]
		}
		result[i] = temp
	}

	return result
}
