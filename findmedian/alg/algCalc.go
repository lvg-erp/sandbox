package alg

import "sort"

type MedianFinder struct {
	numbers []int
}

func NewMedianFinder() *MedianFinder {
	return &MedianFinder{numbers: []int{}}
}

func (t *MedianFinder) AddNum(num int) {
	t.numbers = append(t.numbers, num)
}

func (t *MedianFinder) FindMedian() float64 {
	sort.Ints(t.numbers)
	n := len(t.numbers)
	if n%2 == 0 {
		return float64(t.numbers[n/2-1]+t.numbers[n/2]) / 2.0
	} else {
		return float64(t.numbers[n/2])
	}
}
