package main

import (
	"fmt"
	"sync"
)

type AnyStruct struct {
	safe map[string]int
	mu   sync.Mutex
}

func NewAnyStruct(m map[string]int) *AnyStruct {
	return &AnyStruct{
		safe: m,
	}
}

func main() {

	var wg sync.WaitGroup
	flag := make(chan bool)
	mapSafe := make(map[string]int)
	mp := NewAnyStruct(mapSafe)

	wg.Add(2)
	go func() {
		defer wg.Done()
		mp.Set("a", 1)
		mp.Set("b", 100)
		mp.Set("c", 3)
		fmt.Println(mp)
		flag <- true
		close(flag)
	}()

	go func() {
		defer wg.Done()
		<-flag
		if val, ok := mp.Get("b"); ok {
			fmt.Printf("val %d for key %s\n", val, "b")
		} else {
			fmt.Println("b not found")
		}
	}()

	wg.Wait()

	if ok := mp.Delete("a"); ok {
		fmt.Printf("key1 deleted %s\n", "a")
	} else {
		fmt.Println("a not found")
	}

}

func (a *AnyStruct) Set(Key string, Value int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.safe[Key] = Value
}

func (a *AnyStruct) Get(Key string) (int, bool) {

	a.mu.Lock()
	defer a.mu.Unlock()
	if val, ok := a.safe[Key]; ok {
		return val, true
	}
	return 0, false
}

func (a *AnyStruct) Delete(Key string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, ok := a.safe[Key]; ok {
		delete(a.safe, Key)
		return true
	} else {
		return false
	}

}
