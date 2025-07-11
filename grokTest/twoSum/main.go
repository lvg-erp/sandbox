package main

import "fmt"

func main() {

	const target = 9
	nums := []int{3, 11, 6, 15}
	res := twoSum(nums, target)
	fmt.Println(res)

}

func twoSum(nums []int, target int) []int {
	seen := map[int]int{}
	result := make([]int, 0, len(nums))

	for i, num := range nums {
		complement := target - num
		if j, ok := seen[complement]; ok {
			result = append(result, j, i)
		}
		seen[num] = i
	}

	return result
}
