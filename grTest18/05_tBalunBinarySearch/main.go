package main

import (
	"maps"
	"slices"
)

func main() {

}

// 1 - работает только с отсортированными массивами
func search(nums []int, target int) int {

	left := 0
	right := len(nums) - 1

	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		}
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}

	}

	return -1
}

// 2
func searchInsert(nums []int, target int) int {
	left := 0
	right := len(nums) - 1
	for left <= right {
		mid := left + (right-left)/2
		if nums[mid] == target {
			return mid
		}
		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	return left

}

// 3
func mySqrt(x int) int {
	if x <= 2 {
		return x
	}

	left := 0
	right := x

	for left <= right {
		mid := (left + right) / 2
		if mid*mid == x {
			return mid
		}
		if mid*mid < x {
			left = mid + 1
		} else {
			right = mid - 1
		}

	}

	return right

}

// второй вариант с защитой от переполнения
// если число больше 2 миллиардов
func mySqrt_OverfillProtection(x int) int {
	if x <= 1 {
		return x
	}
	left, right := 0, x
	for left <= right {
		mid := left + (right-left)/2
		if mid <= x/mid { // безопасно от переполнения
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return right
}

// 4
func isPerfectSquare(num int) bool {
	if num < 2 {
		return true
	}

	left := 1
	right := num

	for left <= right {
		mid := left + (right-left)/2
		square := mid * mid

		if square == num {
			return true
		} else if square < num {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return false
}

// 5
// Неправильный
func searchRange_1(nums []int, target int) []int {
	temp := make(map[int]int)
	result := []int{-1, -1}
	left := 0
	right := len(nums) - 1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			temp[mid] = nums[mid]
		}

		if nums[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}

	if len(temp) > 0 {
		result = result[:0]
		keys := slices.Collect(maps.Keys(temp))

		// 2. Сортируем слайс ключей
		slices.Sort(keys)

		// 3. Итерируемся по отсортированным ключам
		for _, k := range keys {
			result = append(result, k)
			// fmt.Printf("Ключ: %d, Значение: %d\n", k, m[k])
		}

	}

	return result
}

// Правильный способ
func searchRange_2(nums []int, target int) []int {
	if len(nums) == 0 {
		return []int{-1, -1}
	}
	left := findLeft(nums, target)
	if left == -1 {
		return []int{-1, -1}
	}
	right := findRight(nums, target)
	return []int{left, right}
}

func findLeft(nums []int, target int) int {
	l, r := 0, len(nums)-1
	for l <= r {
		mid := l + (r-l)/2
		if nums[mid] < target {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	if l < len(nums) && nums[l] == target {
		return l
	}
	return -1
}

func findRight(nums []int, target int) int {
	l, r := 0, len(nums)-1
	for l <= r {
		mid := l + (r-l)/2
		if nums[mid] <= target {
			l = mid + 1
		} else {
			r = mid - 1
		}
	}
	if r >= 0 && nums[r] == target {
		return r
	}
	return -1
}

// 6
func searchMatrix(matrix [][]int, target int) bool {
	if len(matrix) == 0 || len(matrix[0]) == 0 {
		return false
	}
	m, n := len(matrix), len(matrix[0])
	left, right := 0, m*n-1

	for left <= right {
		mid := left + (right-left)/2
		row := mid / n // виртуальная строка
		col := mid % n // виртуальная колонка

		val := matrix[row][col]

		if val == target {
			return true
		}
		if val < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return false
}

// 7
func findPeakElement(nums []int) int {
	if len(nums) == 0 {
		return 0
	}

	left, right := 0, len(nums)-1

	for left < right {
		mid := left + (right-left)/2

		if nums[mid] < nums[mid+1] {
			left = mid + 1
		} else {
			right = mid
		}
	}

	return left

}

// 8
func shipWithinDays(weights []int, days int) int {
	// result := 0
	low := 0
	hight := 0
	for _, w := range weights {
		low = max(low, w)
		hight += w
	}

	for low < hight {
		mid := low + (hight-low)/2
		if canShip(weights, mid, days) {
			hight = mid
		} else {
			low = mid + 1
		}

	}
	return low
}

func canShip(weights []int, capacity int, days int) bool {
	currentDayWeight := 0 // вес который загрузили в текущий день
	needDays := 1         // сколько дней нужно

	for _, w := range weights {
		if currentDayWeight+w > capacity {
			currentDayWeight = w
			needDays++
			if needDays > days {
				return false
			}

		} else {
			currentDayWeight += w
		}
	}

	return true
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
