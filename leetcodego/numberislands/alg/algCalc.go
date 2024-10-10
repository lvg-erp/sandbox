package alg

func SearchEarth(matrix [][]byte) int {

	if len(matrix) == 0 {
		return 0
	}
	countIsland := 0

	for row := 0; row < len(matrix); row++ {
		for col := 0; col < len(matrix[0]); col++ {
			if matrix[row][col] == '1' {
				countIsland++
				dfs(row, col, matrix)
			}
		}
	}

	return countIsland
}

// Функция рекурсивного обхода
// Принимает текущий столбец и ряд, а также
// матрицу
func dfs(row, col int, matrix [][]byte) {

	rows := len(matrix)
	cols := len(matrix[0])

	if row < 0 || col < 0 || row >= rows || col >= cols || matrix[row][col] == '0' {
		return
	}
	//помечаем пройденное поле
	matrix[row][col] = '0'
	//проверим вправо влево вверх вниз наличие 0 или 1
	if row > 0 && matrix[row-1][col] == '1' {
		dfs(row-1, col, matrix)
	}
	if col > 0 && matrix[row][col-1] == '1' {
		dfs(row, col-1, matrix)
	}
	if row < rows-1 && matrix[row+1][col] == '1' {
		dfs(row+1, col-1, matrix)
	}
	if col < cols-1 && matrix[row][col+1] == '1' {
		dfs(row, col+1, matrix)
	}

}
