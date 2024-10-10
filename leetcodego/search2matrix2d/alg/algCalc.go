package alg

func SearchInMatrixLine(matrix [][]int, target int) bool {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return false
	}

	row, col := len(matrix), len(matrix[0])
	for i := 0; i < row; i++ {
		for j := 0; j < col; j++ {
			if matrix[i][j] == target {
				return true
			}
		}
	}

	return false

}

func SearchInMatrix(matrix [][]int, target int) bool {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return false
	}

	array := matrix2array(matrix)
	/// 1,4,5,6,7
	start := 0
	end := len(array) - 1
	for start <= end {
		//либо mid := start + (end - start) / 2
		mid := (end - start) / 2
		curr := start + mid
		if array[curr] > target {
			end = mid - 1
		} else if array[curr] < target {
			start = mid + 1
		} else {
			return true
		}
	}

	return false
}

func matrix2array(matrix [][]int) []int {
	var result []int

	for _, row := range matrix {
		for _, value := range row {
			result = append(result, value)
		}
	}

	return result

}
