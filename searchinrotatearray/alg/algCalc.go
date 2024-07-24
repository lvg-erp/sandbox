package alg

func Search(nums []int, target int) bool {
	n := len(nums)
	if n == 0 {
		return false
	}
	end := n - 1
	start := 0
	for start < end {
		mid := start + (end-start)/2
		if nums[mid] == target {
			return true
		}
		if !isBinarySearchHelpful(nums, start, nums[mid]) {
			start++
			continue
		}
		pivotArray := existsInFirst(nums, start, nums[mid])
		targetArray := existsInFirst(nums, start, target)

		if pivotArray != targetArray {
			if pivotArray {
				start = mid + 1
			} else {
				end = mid - 1
			}
		} else {
			if nums[mid] < target {
				start = mid + 1
			} else {
				end = mid - 1
			}
		}
	}

	return false
}

func isBinarySearchHelpful(nums []int, start, element int) bool {
	return nums[start] != element
}

func existsInFirst(nums []int, start, element int) bool {
	return nums[start] <= element
}
