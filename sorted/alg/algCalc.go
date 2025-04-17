package alg

import "math/rand"

func CreateArr() (arr []int) {
	//size := 5000
	size := 5
	arr = make([]int, size)
	for i := 0; i < size; i++ {
		arr[i] = rand.Intn(size)
	}

	return
}

// Сотрировка пузырьковая сложность O^2 (два цикла - сложность в квадрате )
func BubbleSort(arr []int) []int {
	for i := 0; i < len(arr); i++ {
		for j := i + 1; j < len(arr); j++ {
			if arr[i] > arr[j] {
				arr[i], arr[j] = arr[j], arr[i]
			}
		}
	}

	return arr
}

// Сортировка выбором помещаем мин элемент в самое начало сложность O^2 (два цикла - сложность в квадрате )
func SelectionSort(arr []int) []int {
	for i := 0; i < len(arr)-1; i++ {
		min := i
		for j := i + 1; j < len(arr); j++ {
			if arr[j] < arr[min] {
				min = j
			}
		}
		arr[i], arr[min] = arr[min], arr[i]
	}

	return arr
}

// Сортировка вставками сложность O^2 (два цикла - сложность в квадрате )
func InsertionSort(arr []int) []int {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}

	return arr

}

// Быстрая сортировка рекурсия сложность О(n X log n)
func QuickSort(arr []int) []int {
	if len(arr) < 2 {
		return arr
	}

	pivot := arr[0] // выбираем опрный элемент
	var less, greater []int
	for _, num := range arr[1:] {
		if num <= pivot {
			less = append(less, num)
		} else {
			greater = append(greater, num)
		}
	}

	result := append(QuickSort(less), pivot)       // элементы меньше опрного помещаем до
	result = append(result, QuickSort(greater)...) // элементы больше опрного помещаем после (три точки - это распаковка)

	return result
}

// Сортировка слиянием сложность О(n X log n)
func MergeSort(arr []int) []int {

	if len(arr) < 2 {
		return arr
	}

	mid := len(arr) / 2
	left := MergeSort(arr[:mid])
	right := MergeSort(arr[mid:])
	return merge(left, right)

}

func merge(left, right []int) []int {
	var merged []int
	for len(left) > 0 && len(right) > 0 {
		if left[0] <= right[0] {
			merged = append(merged, left[0])
			left = left[1:]
		} else {
			merged = append(merged, right[0])
			right = right[1:]
		}
	}

	merged = append(merged, left...)
	merged = append(merged, right...)
	return merged
}
