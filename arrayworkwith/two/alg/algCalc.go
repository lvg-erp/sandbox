package alg

func MoveZero2End(nums []int) {
	lastNonZeroFoundAt := 0
	for cur := 0; cur < len(nums); cur++ {
		if nums[cur] != 0 {
			nums[lastNonZeroFoundAt], nums[cur] = nums[cur], nums[lastNonZeroFoundAt]
			lastNonZeroFoundAt++
		}
	}
}
