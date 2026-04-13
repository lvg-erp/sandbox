package main

import "fmt"

func main() {

	nums := []int{1, 2, 3, 4, 5, 6, 7, 7, 7, 7, 8, 8, 9, 9}
	target := 7

	res := searchStartEndWithDubl(nums, target)

	fmt.Println(res)

}

func searchStartEndWithDubl(nums []int, target int) []int {
	var result []int
	lidx := searchLeft(nums, target)
	result = append(result, lidx)
	ridx := searchRight(nums, target)
	result = append(result, ridx)

	return result
}

func searchLeft(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left < right {
		mid := left + (right-left)/2
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid
		}
	}
	if nums[left] == target {
		return left
	}
	return -1
}

func searchRight(nums []int, target int) int {
	left, right := 0, len(nums)-1
	for left < right {
		mid := left + (right-left+1)/2
		fmt.Println(mid)
		if nums[mid] > target {
			right = mid - 1
		} else {
			left = mid
		}
	}
	if nums[left] == target {
		return left
	}
	return -1
}

// Вариант от ИИ
//func searchRange(nums []int, target int) []int {
//	if len(nums) == 0 {
//		return []int{-1, -1}
//	}
//
//	left := binarySearch(nums, target, true)
//	if left == -1 {
//		return []int{-1, -1}
//	}
//	right := binarySearch(nums, target, false)
//
//	return []int{left, right}
//}
//
//// one шаблон для обоих случаев
//func binarySearch(nums []int, target int, isLeft bool) int {
//	l, r := 0, len(nums)-1
//	result := -1
//
//	for l <= r {
//		mid := l + (r-l)/2
//
//		if nums[mid] == target {
//			result = mid
//			if isLeft {
//				r = mid - 1   // ищем левее
//			} else {
//				l = mid + 1   // ищем правее
//			}
//		} else if nums[mid] < target {
//			l = mid + 1
//		} else {
//			r = mid - 1
//		}
//	}
//	return result
//}
// ЭТО НЕВЕРНОЕ РЕШЕНИЕ!!!!!!!!!!!!!
//func searchStartEndWithDubl(nums []int, target int) []int {
//	var result []int
//	lidx := searchLeft(nums, target)
//	result = append(result, lidx)
//	ridx := searchRight(nums, target)
//	result = append(result, ridx)
//
//	return result
//}
//
//func searchLeft(nums []int, target int) int {
//	left := 0
//	right := len(nums)
//
//	for left <= right {
//		mid := left + (right-left)/2
//		if nums[mid] == target {
//			return mid
//		}
//
//		if nums[left] < target {
//			left = mid + 1
//		} else {
//			right = mid - 1
//		}
//	}
//
//	return -1
//}
//
//func searchRight(nums []int, target int) int {
//	left := 0
//	right := len(nums) - 1
//
//	for right >= left {
//		mid := right - left
//		fmt.Println(mid)
//		if nums[mid] == target {
//			return mid
//		}
//
//		if nums[right] > target {
//			right = mid - 1
//		} else {
//			left = mid + 1
//		}
//
//	}
//
//	return -1
//}
