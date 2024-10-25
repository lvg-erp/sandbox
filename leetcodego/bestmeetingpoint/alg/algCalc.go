package alg

import "math"

func MinTotalDistance(grid [][]int) int {
	minDistance := math.MaxInt32
	for row := 0; row < len(grid); row++ {
		for col := 0; col < len(grid[0]); col++ {
			distance := search(grid, row, col)
			if distance < minDistance {
				minDistance = distance
			}
		}
	}

	return minDistance
}

func search(grid [][]int, row int, col int) int {
	type Point struct {
		r, c, d int
	}

	q := []Point{{row, col, 0}}
	m := len(grid)
	n := len(grid[0])

	visited := make([][]bool, m)
	for i := range visited {
		visited[i] = make([]bool, n)
	}
	totalDistance := 0

	for len(q) > 0 {
		point := q[0]
		q = q[1:]
		r, c, d := point.r, point.c, point.d
		if r < 0 || c < 0 || r >= m || c >= n || visited[r][c] {
			continue
		}

		if grid[r][c] == 1 {
			totalDistance += d
		}

		visited[r][c] = true

		q = append(q, Point{r + 1, c, d + 1})
		q = append(q, Point{r - 1, c, d + 1})
		q = append(q, Point{r, c + 1, d + 1})
		q = append(q, Point{r, c - 1, d + 1})
	}

	return totalDistance

}
