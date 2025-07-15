package main

import (
	"testing"
)

func TestTwoSum(t *testing.T) {
	test := []struct {
		name   string
		nums   []int
		target int
		want   []int
	}{
		{name: "target_9", nums: []int{2, 7, 11, 15}, target: 9, want: []int{0, 1}},
	}

	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			got := twoSum(tt.nums, tt.target)
			if len(got) != 2 {
				t.Errorf("Expected 2 indices, got %d", len(got))
				return
			}
			i, j := got[0], got[1]
			if i == j || tt.nums[i]+tt.nums[j] != tt.target {
				t.Errorf("Invalid indices %v: sum=%d, want target=%d", got, tt.nums[i]+tt.nums[j], tt.target)
			}
		})
	}

}
