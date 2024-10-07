package alg

type Solution struct {
	N     int
	Graph [][]int
}

func (s *Solution) FindCelebrity() int {
	celebrityCandidate := 0
	for i := 1; i < s.N; i++ {
		if s.knows(celebrityCandidate, i) {
			celebrityCandidate = i
		}
	}
	if s.isCelebrity(celebrityCandidate) {
		return celebrityCandidate
	}
	return -1
}

func (s *Solution) isCelebrity(i int) bool {
	for j := 0; j < s.N; j++ {
		if i == j {
			continue
		}
		if s.knows(i, j) || !s.knows(j, i) { // Celebrity knows someone or someone doesn't know celebrity
			return false
		}
	}
	return true
}

func (s *Solution) knows(a, b int) bool {
	return s.Graph[a][b] == 1
}
