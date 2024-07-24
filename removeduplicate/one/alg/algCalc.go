package alg

func RemoveDuplicates(nums []int) (res int, arr []int) {
	j := 0

	for i := 0; i < len(nums); i++ {
		if j < 2 || nums[i] > nums[j-2] {
			nums[j] = nums[i]
			j++
		}
	}

	return j, nums
}
