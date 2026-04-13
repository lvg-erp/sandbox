package main

import "fmt"

func main() {
	// 2
	//nums := []int{1, 2, 3, 4, 5, 5, 5, 7, 8, 9, 9}
	//target := 10
	//
	//m := searchPointToInsert(nums, target)
	//fmt.Println(m)

	// 3

	nums := []int{5, 5, 5, 7, 8, 9, 9, 1, 2, 3, 4}
	target := 3

	m := searchRotateArray(nums, target)
	fmt.Println(m)
}

// 2
func searchPointToInsert(nums []int, target int) map[string]int {

	result := make(map[string]int)

	left := 0
	right := len(nums) - 1

	for left < right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			result["index for insert "] = mid
			return result
		}
		// Этот кусок не нужен
		//if nums[mid-1] < target && nums[mid+1] > target {
		//	result["index for insert "] = mid
		//}

		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}

	}

	result["index for insert "] = left

	return result
}

// 3

func searchRotateArray(nums []int, target int) int {
	left, right := 0, len(nums)-1

	for left < right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		}

		// проверяем на наличие дублей
		if nums[left] == nums[mid] && nums[mid] == nums[right] {
			left++
			right--
			continue
		}
		// Вариант 1 - левая часть отсортирована
		if nums[left] <= nums[mid] {
			if nums[left] <= target && target <= nums[mid] {
				right = mid - 1
			} else {
				left = mid + 1
			}
		} else { // Вариант 2 - правая часть отсортирована
			if nums[mid] <= target && target <= nums[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}

	}

	return -1
}
