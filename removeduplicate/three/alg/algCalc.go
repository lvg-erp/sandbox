package alg

func ContainsNearbyDuplicate(nums []int, k int) bool {
	set := make(map[int]struct{})
	for i, num := range nums {
		if _, exists := set[num]; exists {
			return true
		}
		set[num] = struct{}{}
		if len(set) > k {
			delete(set, nums[i-k])
		}
	}

	return false
}
