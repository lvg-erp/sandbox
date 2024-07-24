package main

import (
	"fmt"
	"sorted/alg"
	"time"
)

func main() {
	arr := alg.CreateArr()
	fmt.Println(arr[:10])

	emptyArr := make([]int, len(arr))
	copy(emptyArr, arr)
	start := time.Now()
	//alg.BubbleSort(emptyArr)
	//alg.SelectionSort(emptyArr)
	//alg.InsertionSort(emptyArr)
	//alg.QuickSort(emptyArr)
	alg.MergeSort(emptyArr)
	duration := time.Since(start)

	//fmt.Println("Пузырьковая сортировка O(n^2) занимает: ", duration)
	//fmt.Println("Сортировка выбором O(n^2) занимает: ", duration)
	//fmt.Println("Сортировка вставками O(n^2) занимает: ", duration)
	//fmt.Println("Быстрая сортировка рекурсия сложность О(n X log n) занимает: ", duration)
	fmt.Println("Сортировка слиянием сложность О(n X log n) занимает: ", duration)
}
