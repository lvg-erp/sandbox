package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
)

//Обрабатывать 10 документов TmcDocument, отправляя их пачками по 3 документа в 3 горутины.
//Это означает, что одновременно будут работать не более 3 горутин,
//каждая из которых обрабатывает группу из 3 документов (или меньше, если осталось меньше документов).
//Для этого мы будем использовать ограниченный пул горутин (как в оригинальной задаче 3) и разделим документы на пачки.

// TODO: структуры вынести в отдельный модуль
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

type TaskResult struct {
	DocumentGUID string
	Success      bool
	Error        error
}

type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// заглушка для клиента
type StubHTTPClient struct{}

func (c *StubHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"status": "success"}`))),
	}, nil
}

func processTasksWithPool(documents TmcDocumentsResponse, baseUrl string, client HTTPClient) ([]TaskResult, error) {
	const maxWorkers = 3 // количество горутин
	const bathSize = 3   // пакеты по 3 штуки

	if len(documents) == 0 {
		return nil, fmt.Errorf("no documents to process")
	}
	var wg sync.WaitGroup
	result := make([]TaskResult, len(documents))
	errChan := make(chan error, len(documents))
	semaphore := make(chan struct{}, maxWorkers)
	for i := 0; i < len(documents); i += bathSize {
		end := i + bathSize
		if end > len(documents) {
			end = len(documents)
		}
		bath := documents[i:end]
		wg.Add(1)
		semaphore <- struct{}{}
		go func(bath []TmcDocument, startIdx int) {
			defer wg.Done()
			defer func() { <-semaphore }()

			for idx, doc := range bath {
				err := processSingleTask(doc, baseUrl, client)
				result[startIdx+idx] = TaskResult{
					DocumentGUID: doc.TmcDocumentGUID,
					Success:      err == nil,
					Error:        err,
				}
				if err != nil {
					log.Printf("Error processing document %s: %w", doc.TmcDocumentGUID, err)
					errChan <- err
				}
			}
		}(bath, i)
	}
	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)

	}

	if len(errors) > 0 {
		return result, fmt.Errorf("errors occurred processing: %v", errors)
	}

	return result, nil

}

func processSingleTask(doc TmcDocument, baseUrl string, client HTTPClient) error {
	requestBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("json.MArshall failed for document %s: %w", doc.TmcDocumentGUID, err)
	}

	requestURL := fmt.Sprintf("%s/%s", baseUrl, doc.TmcDocumentGUID)
	resp, err := client.Post(requestURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("HTTP request failed for document %s: %w", doc.TmcDocumentGUID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for document %s: %w", doc.TmcDocumentGUID, err)
	}

	return nil
}

func main() {
	jsonData, err := os.ReadFile("documents.json")
	if err != nil {
		log.Fatalf("Failed to read documents.json: %v", err)
	}
	var documents TmcDocumentsResponse
	if err := json.Unmarshal(jsonData, &documents); err != nil {
		log.Fatalf("Failed to unmarshaling jsonData: %v", err)
	}

	client := &StubHTTPClient{}

	results, err := processTasksWithPool(documents, "http://stub.api", client)
	if err != nil {
		log.Printf("processTaskWithPoll failed: %v", err)
	}

	for _, res := range results {
		if !res.Success {
			log.Printf("Expected successful processing for document %s, got error: %v", res.DocumentGUID, err)
		}
	}

	log.Println("All documents processed successfully:")
	for _, res := range results {
		log.Printf("Document %s: Success = %v, Error = %v", res.DocumentGUID, res.Success, res.Error)
	}

}
