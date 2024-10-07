package alg

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MaximalSquare(matrix [][]byte) int {
	rows := len(matrix)
	cols := 0

	if rows > 0 {
		cols = len(matrix[0])
	}

	dp := make([]int, cols+1)
	maxqlen := 0
	prev := 0
	for i := 1; i <= rows; i++ {
		for j := 1; j <= cols; j++ {
			temp := dp[j]
			if matrix[i-1][j-1] == '1' {
				dp[j] = min(dp[j-1], min(prev, dp[j])) + 1
				maxqlen = max(maxqlen, dp[j])
			} else {
				dp[j] = 0
			}

			prev = temp
		}
	}

	return maxqlen * maxqlen
}

func StringToByte(arr_str [][]string) [][]byte {
	byteSlice := make([][]byte, len(arr_str))

	for i, row := range arr_str {
		byteSlice[i] = make([]byte, len(row))
		for j, str := range row {
			byteSlice[i][j] = str[0] // Берем первый байт строки (символа)
		}
	}

	return byteSlice
}
