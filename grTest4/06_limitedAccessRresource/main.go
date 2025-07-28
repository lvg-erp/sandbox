package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const (
	workers   = 100
	maxAccess = 3
)

// будем пытаться читать файл в структуру
type Tmc struct {
	TmcGUID           string `json:"TmcGUID"`
	TmcName           string `json:"TmcName"`
	UnitOfMeasurement string `json:"UnitOfMeasurement"`
	Count             int    `json:"Count"`
	Enable            bool   `json:"Enable"`
}

type TmcDocument struct {
	TmcDocumentGUID   string `json:"TmcDocumentGUID"`
	TmcDocumentNumber string `json:"TmcDocumentNumber"`
	CreateTime        int64  `json:"CreateTime"`
	Status            string `json:"Status"`
	AcsessTime        int64  `json:"AcsessTime"`
	ServiceGUID       string `json:"ServiceGUID"`
	EmployerGUID      string `json:"EmployerGUID"`
	DriverGUID        string `json:"DriverGUID"`
	TmcList           []Tmc  `json:"TmcList"`
}

type TmcDocumentsResponse []TmcDocument

func main() {
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxAccess)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		//???????????semaphore <- struct{}{}
		go func(idx int) {
			defer wg.Done()
			defer func() {
				fmt.Printf("Goroutine %d released resource at %v\n", idx, time.Now().Format(time.RFC3339))
				<-semaphore
			}()
			semaphore <- struct{}{}
			fmt.Printf("Goroutine %d acquired resource at %v\n", idx, time.Now().Format(time.RFC3339))
			jsonData, err := os.ReadFile("documents.json")
			if err != nil {
				log.Printf("Failed to read documents.json: %v", err)
			}
			var documents TmcDocumentsResponse
			if err := json.Unmarshal(jsonData, &documents); err != nil {
				log.Printf("Failed to unmarshaling jsonData: %v", err)
			}

			time.Sleep(1 * time.Second)
		}(i)
	}

	wg.Wait()

}
