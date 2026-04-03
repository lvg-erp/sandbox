package main

import (
	"regexp"
	"sort"
	"strings"
)

func main() {

}

// 1
func reverseString(s []byte) {
	left, right := 0, len(s)-1

	for left < right {
		s[left], s[right] = s[right], s[left]
		left++
		right--
	}

}

// 2
func isPalindrome(s string) bool {

	if len(s) < 2 {
		return true
	}
	// убираем пробелы и знаки припинания.
	var re = regexp.MustCompile(`[[:punct:]]|[[:space:]]`)
	// приводим все к нижнему регистру.
	str := re.ReplaceAllString(strings.ToLower(s), "")
	left, right := 0, len(str)-1
	for left < right {
		if str[left] == str[right] {
			left++
			right--
		} else {
			return false
		}
	}

	return true

}

// 3
// первый способ
func merge_1(nums1 []int, m int, nums2 []int, n int) {
	copy(nums1[m:], nums2)

	sort.Ints(nums1)
}

// второй способ

func merge_2(nums1 []int, m int, nums2 []int, n int) {
	i := m - 1     // конец реальной части nums1
	j := n - 1     // конец nums2
	k := m + n - 1 // конец nums1

	for j >= 0 {
		if i >= 0 && nums1[i] > nums2[j] {
			nums1[k] = nums1[i]
			i--
		} else {
			nums1[k] = nums2[j]
			j--
		}
		k--
	}
}

// 4
func intersection(nums1 []int, nums2 []int) []int {
	mp := make(map[int]struct{})
	var result []int
	for _, v := range nums1 {
		mp[v] = struct{}{}
	}

	for _, v := range nums2 {
		if _, ok := mp[v]; ok {
			result = append(result, v)
			delete(mp, v)
		}
	}

	return result

}

// 5
func sortedSquares(nums []int) []int {
	ln := len(nums)
	result := make([]int, ln)
	left := 0
	right := ln - 1
	k := ln - 1
	for left <= right {
		lSq := nums[left] * nums[left]
		rSq := nums[right] * nums[right]
		if lSq > rSq {
			result[k] = lSq
			left++
		} else {
			result[k] = rSq
			right--
		}
		k--
	}

	return result

}
