package alg

import (
	"golang.org/x/exp/rand"
	"time"
)

func string2Map(input string) map[rune]int {
	charCount := make(map[rune]int)

	// Подсчет вхождений каждого символа
	for _, char := range input {
		charCount[char]++
	}

	return charCount
}

func SearchUnicodeCharacter(input string) int {

	hm := string2Map(input)

	for ind, char := range input {
		if hm[char] == 1 {
			return ind
		}
	}

	//for char, ind := range hm {
	//	if hm[char] == 1 {
	//		return ind
	//	}
	//}

	return -1

}

//func String2Map(input string) map[int]rune {
//	out := make(map[int]rune)
//	// Подсчет вхождений каждого символа
//	for i, char := range input {
//		out[i] = char
//	}
//
//	return out
//}

type Solution struct {
	original []int
	array    []int
}

func NewSolution(nums []int) *Solution {
	original := make([]int, len(nums))
	copy(original, nums)
	return &Solution{
		original: original,
		array:    append([]int(nil), nums...),
	}
}

func (s *Solution) Reset() []int {
	s.array = append([]int(nil), s.original...)
	return s.original
}

func (s *Solution) Shuffle() []int {
	rand.Seed(uint64(time.Now().UnixNano()))
	for i := range s.array {
		randIdex := rand.Intn(len(s.array)-i) + i
		s.array[i], s.array[randIdex] = s.array[randIdex], s.array[i]
	}

	return s.array
}
