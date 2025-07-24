package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"sync"
	"time"
)

const workers = 4

func main() {

	var wg sync.WaitGroup
	var mu sync.Mutex
	mapCount := make(map[string]int)

	//var ars []string
	ch := make(chan string, 200)
	go func() {
		for i := 0; i < 100; i++ {
			b := generateRandomBytes()
			ch <- string(b)
			//ar := generateStringArray(s)
			//ars = append(ars, ar...)
		}
		close(ch)
	}()

	for y := 0; y < workers; y++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for c := range ch {
				mu.Lock()
				mapCount[c]++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	fmt.Println(mapCount)

}

//func generateRandomBytes() []byte {
//	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//	const length = 6
//	rand.Seed(uint64(time.Now().UnixNano()))
//
//	b := make([]byte, length)
//	for i := range b {
//		b[i] = charset[rand.Intn(len(charset))]
//	}
//	return b
//}

//для генерации повторяющих ся ключей

func generateRandomBytes() []byte {
	fixedSet := []string{
		"K9xP4m",
		"tR7wQ2",
		"Z8nL5j",
		"aB1cD2",
		"Xy9zW4",
		"Pq3Rs5",
		"Lm8Nk7",
		"Jh2Tv9",
		"Fg6Yx1",
		"De4Uw8",
	}
	rand.Seed(uint64(time.Now().UnixNano()))
	return []byte(fixedSet[rand.Intn(len(fixedSet))])
}

//func generateStringArray(in string) []string {
//	var out []string
//	out = append(out, in)
//	return out
//}

//func generateString() string {
//	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//	const length = 6
//	rand.Seed(uint64(time.Now().UnixNano()))
//
//	b := make([]byte, length)
//	for i := range b {
//		b[i] = charset[rand.Intn(len(charset))]
//	}
//	return string(b)
//
//}
