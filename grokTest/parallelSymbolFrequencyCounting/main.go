package main

import (
	"fmt"
	"sort"
	"sync"
)

func main() {
	sts := []string{"hello", "world", "change"}

	res := countFrequency(sts, 2)
	slf := mapRuneToMapString(res)
	srt := sortingMapByValue(slf)

	fmt.Println(srt)

}

func mapRuneToMapString(res map[rune]int) map[string]int {
	symbolF := make(map[string]int)
	for key, i := range res {
		symbolF[string(key)] += i
	}

	return symbolF

}

func countSymbolsRuneOnString(sts []string) map[rune]int {
	result := make(map[rune]int)
	for _, st := range sts {
		for _, char := range st {
			result[char]++
		}
	}

	return result
}

func countFrequency(sts []string, workers int) map[rune]int {

	if workers <= 0 {
		workers = 1
	}

	resultCh := make(chan map[rune]int, workers)
	finalFreq := make(map[rune]int)
	var wg sync.WaitGroup
	chunkSize := (len(sts) + workers - 1) / workers
	for i := 0; i < workers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(sts) {
			end = len(sts)
		}
		if start < len(sts) {
			wg.Add(1)
			go func(chunk []string) {
				defer wg.Done()
				freg := countSymbolsRuneOnString(chunk) // убрал sts иначе каждая горутина обрабатывает весь слайс, а не свою часть. Это приводит к дублированию результатов.
				resultCh <- freg
			}(sts[start:end])
		}

	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for ch := range resultCh {
		for r, count := range ch {
			finalFreq[r] += count
		}
	}

	return finalFreq
}

type pair struct {
	Key   string
	Value int
}

func sortingMapByValue(unsortedMap map[string]int) []pair {

	sortedMap := make(map[string]int)
	var pairs []pair
	for k, v := range unsortedMap {
		pairs = append(pairs, pair{k, v})
	}

	sort.Slice(pairs, func(i, j int) bool {
		//TODO: сортировка вынести в отдельную функцию
		if pairs[i].Value == pairs[j].Value {
			return pairs[i].Key > pairs[j].Key
		}
		return pairs[i].Value > pairs[j].Value
	})

	for _, p := range pairs {
		sortedMap[p.Key] = p.Value
	}
	// TODO: мап нельзя отсортировать!
	//fmt.Println("Отсортированная map:")
	//for k, v := range sortedMap {
	//	fmt.Printf("%s: %d\n", k, v)
	//}

	return pairs

}
