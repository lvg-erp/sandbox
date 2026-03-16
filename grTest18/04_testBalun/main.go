package main

import (
	"fmt"
	"sort"
)

func main() {
	//sIn := "foo"
	//tIn := "bar"
	//res := isIsomorphic(sIn, tIn)
	//
	//fmt.Println(res)

	//s := "MCMXCIV"
	//res := romanToInt(s)
	strs := []string{"eat", "tea", "tan", "ate", "nat", "bat"}
	res := groupAnagrams(strs)
	fmt.Println(res)
}

// 1
func twoSum(nums []int, target int) []int {
	seen := make(map[int]int)
	for i, num := range nums {
		if j, ok := seen[target-num]; ok {
			return []int{j, i}
		}
		seen[num] = i
	}

	return nil

}

// 2
func isIsomorphic(s string, t string) bool {
	if len(s) != len(t) {
		return false
	}

	sT := make(map[byte]byte)
	tU := make(map[byte]bool)

	for i := 0; i < len(s); i++ {
		cs, ct := s[i], t[i]
		if mapped, exists := sT[cs]; exists {
			if mapped != ct {
				return false
			}
		} else {
			if tU[ct] {
				return false
			}
		}
		sT[cs] = ct
		tU[ct] = true
	}

	return true

}

// 3
func romanToInt(s string) int {
	roman := map[rune]int{
		'I': 1, 'V': 5, 'X': 10, 'L': 50,
		'C': 100, 'D': 500, 'M': 1000,
	}

	prev := 0
	total := 0

	for i := len(s) - 1; i >= 0; i-- {

		val := roman[rune(s[i])]

		if val < prev {
			total -= val
		} else {
			total += val
		}

		prev = val
	}

	return total

}

// 4
func isAnagram(s string, t string) bool {
	if len(s) != len(t) {
		return false
	}
	sCheck := make(map[rune]bool)
	for i := 0; i < len(s)-1; i++ {
		sCheck[rune(s[i])] = true
	}

	for y := 0; y < len(t)-1; y++ {
		if ok := sCheck[rune(t[y])]; !ok {
			return false
		}
	}

	return true
}

// 5
func groupAnagrams(strs []string) [][]string {
	mGroups := make(map[string][]string)

	for _, s := range strs {
		sorted := sortString(s)
		mGroups[sorted] = append(mGroups[sorted], s)
	}

	result := make([][]string, 0, len(mGroups))
	for _, mg := range mGroups {
		result = append(result, mg)
	}

	return result

}

func sortString(s string) string {
	runes := []rune(s)
	sort.Slice(runes, func(i, j int) bool {
		return runes[i] < runes[j]
	})
	return string(runes)
}
