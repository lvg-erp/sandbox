package alg

import "sort"

func GetSkyline(buildings [][]int) [][]int {
	edgeSet := make(map[int]struct{})
	for _, building := range buildings {
		edgeSet[building[0]] = struct{}{}
		edgeSet[building[1]] = struct{}{}
	}

	edges := make([]int, 0, len(edgeSet))

	for edge := range edgeSet {
		edges = append(edges, edge)
	}
	sort.Ints(edges)
	edgeIndexMap := make(map[int]int)
	for i, edge := range edges {
		edgeIndexMap[edge] = i
	}

	heights := make([]int, len(edges))
	for _, building := range buildings {
		left, right, heigh := building[0], building[1], building[2]
		leftIndex, rightIndex := edgeIndexMap[left], edgeIndexMap[right]
		for idx := leftIndex; idx < rightIndex; idx++ {
			if heights[idx] < heigh {
				heights[idx] = heigh
			}
		}
	}

	var answer [][]int
	for i, currHeight := range heights {
		currPos := edges[i]
		if len(answer) == 0 || answer[len(answer)-1][1] != currHeight {
			answer = append(answer, []int{currPos, currHeight})
		}
	}

	return answer
}
