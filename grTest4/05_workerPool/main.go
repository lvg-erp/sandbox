package main

import (
	"fmt"
	"golang.org/x/exp/rand"
	"strings"
	"sync"
	"time"
)

const workers = 5

type Job struct {
	//mu   sync.Mutex
	ID   int
	Data string
}

type Result struct {
	JobID int
	Data  string
}

//В цикле for jb := range chJob переменная jb является копией структуры Job,
//полученной из канала chJob. Поскольку Job содержит поле mu sync.Mutex,
//это вызывает ошибку компиляции: Range variable 'jb' copies the lock: type 'Job' contains 'sync.Mutex' which is 'sync.Locker'.
//Копирование мьютекса недопустимо, так как оно нарушает состояние блокировки.

func (j *Job) writeDataJob(idx int, data string) {
	//j.mu.Lock()
	//defer j.mu.Unlock()
	j.ID = idx
	j.Data = data
}

func NewJob() *Job { return &Job{} }

func main() {

	job := NewJob()
	var wg sync.WaitGroup
	var mu sync.Mutex
	chJob := make(chan *Job, 20)
	chResult := make(chan Result, 20)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < 20; i++ {
			mu.Lock()
			dataBody := generateString()
			job.writeDataJob(i, dataBody)
			mu.Unlock()
			chJob <- job
		}
		close(chJob)
	}()

	for y := 0; y < workers; y++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for jb := range chJob {
				//jb.mu.Lock()
				result := Result{
					JobID: jb.ID,
					Data:  strings.ToUpper(jb.Data),
				}
				//jb.mu.Unlock()
				chResult <- result
			}
		}()
	}

	go func() {
		wg.Wait()
		close(chResult)
	}()

	for f := range chResult {
		fmt.Println(f.JobID, f.Data)
	}
}

func generateString() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	rand.Seed(uint64(time.Now().UnixNano()))

	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)

}
