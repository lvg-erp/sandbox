package main

import (
	"competitivehttp/alg"
	"fmt"
	"sync"
)

func main() {
	var urls = []string{
		"https://google.com",
		"https://somesite.com",
		"https://ya.ru",
		"https://dzen.ru",
		"https://youtube.com",
		"http://non-existent.domain.tld",
	}

	results := make(chan string)
	jobs := make(chan string)

	go func() {
		for _, u := range urls {
			jobs <- u
		}
		close(jobs)
	}()

	workers := 4

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case url, ok := <-jobs:
					if !ok {
						return
					}
					alg.FetchAPI(url, results)
				}
			}
		}()
	}

	go func() {
		wg.Wait()
	}()

	for results := range results {
		fmt.Println(results)
	}

}
