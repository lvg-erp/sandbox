package alg

func DeleteDuplicates(arr1, arr2 []string) []string {
	uniqueRes := make(map[string]struct{})
	result := make([]string, 0)
	for _, v1 := range arr1 {
		uniqueRes[v1] = struct{}{}
	}

	//for _, v2 := range arr2 {
	//	if _, ok := uniqueRes[v2]; !ok {
	//		result = append(result, v2)
	//	}
	//}

	for _, v2 := range arr2 {
		uniqueRes[v2] = struct{}{}
	}

	for k := range uniqueRes {
		result = append(result, k)
	}

	return result
}

func MergeArrayWithoutDuplicate(arr1, arr2 []string) []string {

	uniqueRes := make(map[string]struct{})
	result := make([]string, 0)
	for _, v1 := range arr1 {
		uniqueRes[v1] = struct{}{}
	}

	for _, v2 := range arr2 {
		if _, ok := uniqueRes[v2]; !ok {
			result = append(result, v2)
		}
	}

	for _, item := range arr1 {
		found := false
		for _, item2 := range arr2 {
			if item == item2 {
				found = true
				break
			}
		}
		if !found {
			result = append(result, item)
		}
	}

	return result
}

// Линейный поиск индекса элемента
func LinearSearch(arr []int, s int) int {
	for i, v := range arr {
		if v == s {
			return i
		}
	}

	return -1
}

//Бинарный поиск индекса элемента (массив должен быть отсортирован)

func BinarySearch(arr []int, s int) int {
	leftPointer := 0
	rightPointer := len(arr) - 1
	for leftPointer <= rightPointer {
		midPointer := int((leftPointer + rightPointer) / 2)
		midValue := arr[midPointer]

		if midValue == s {
			return midValue
		} else if midValue < s {
			leftPointer = midPointer + 1
		} else {
			rightPointer = midPointer - 1
		}
	}

	return -1

}
