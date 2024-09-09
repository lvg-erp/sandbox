package algCalc

import (
	"container/heap"
	"sort"
)

// 1
type Solution struct{}

func (s *Solution) overlap(interval1, interval2 []int) bool {
	return (interval1[0] >= interval2[0] && interval1[0] < interval2[1]) ||
		(interval2[0] >= interval1[0] && interval2[0] < interval1[1])
}

func (s *Solution) CanAttendMeetings(intervals [][]int) bool {
	for i := 0; i < len(intervals); i++ {
		for j := i + 1; j < len(intervals); j++ {
			if s.overlap(intervals[i], intervals[j]) {
				return false
			}
		}
	}

	return true
}

//2

type MinHeap []int

func (h MinHeap) Len() int           { return len(h) }
func (h MinHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h MinHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(int))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]

	return x
}

func MinCountMeetingRooms(intervals [][]int) int {
	sort.Slice(intervals, func(i, j int) bool {
		return intervals[i][0] < intervals[j][0]
	})

	h := &MinHeap{intervals[0][1]}
	heap.Init(h)

	for i := 1; i < len(intervals); i++ {
		if intervals[i][0] >= (*h)[0] {
			heap.Pop(h)
		}
		heap.Push(h, intervals[i][1])
	}

	return h.Len()
}
