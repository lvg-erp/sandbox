package alg

func arrayContains(arr []int, num int) bool {
	for i := 0; i < len(arr); i++ {
		if arr[i] == num {
			return true
		}
	}

	return false
}

func LongestConsecutive(nums []int) int {
	longestStreak := 0
	for _, num := range nums {
		currentNum := num
		currentStreak := 1
		for arrayContains(nums, currentNum+1) {
			currentNum += 1
			currentStreak += 1
		}
		if currentStreak > longestStreak {
			longestStreak = currentStreak
		}
	}

	return longestStreak
}
