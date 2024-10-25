package alg

func LongestInscreasingPath(matrix [][]int) int {
	dir := [][]int{{0, 1}, {1, 0}, {0, -1}, {-1, 0}}
	m := len(matrix)
	if m == 0 {
		return 0
	}
	n := len(matrix[0])
	outdegree := make([][]int, m+2)
	newMatrix := make([][]int, m+2)
	for i := range outdegree {
		outdegree[i] = make([]int, n+2)
		newMatrix[i] = make([]int, n+2)
	}

	for i := 0; i < m; i++ {
		copy(newMatrix[i+1][1:], matrix[i])
	}

	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			for _, d := range dir {
				if newMatrix[i][j] < newMatrix[i+d[0]][j+d[1]] {
					outdegree[i][j]++
				}
			}
		}
	}

	var leaves [][]int
	for i := 1; i <= m; i++ {
		for j := 1; j <= n; j++ {
			if outdegree[i][j] == 0 {
				leaves = append(leaves, []int{i, j})
			}
		}
	}

	height := 0
	for len(leaves) > 0 {
		height++
		var newLeaves [][]int
		for _, node := range leaves {
			for _, d := range dir {
				x, y := node[0]+d[0], node[1]+d[1]
				if newMatrix[node[0]][node[1]] > newMatrix[x][y] {
					outdegree[x][y]--
					if outdegree[x][y] == 0 {
						newLeaves = append(newLeaves, []int{x, y})
					}
				}
			}
		}
		leaves = newLeaves
	}

	return height
}
