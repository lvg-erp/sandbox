package main

//Реализовать функцию processTasks, которая:
//Принимает список задач (например, TmcDocumentsResponse).
//Обрабатывает каждую задачу в отдельной горутине (например, отправляет HTTP-запрос или выполняет другую операцию).
//Использует sync.WaitGroup для ожидания завершения всех горутин.
//Использует каналы для сбора ошибок или результатов.
//Возвращает результаты обработки и/или ошибки.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
)

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

// Интерфейс для HTTP-клиента
type HTTPClient interface {
	Post(url, contentType string, body io.Reader) (*http.Response, error)
}

// Заглушка клиента
type StubHTTPClient struct{}

func (c *StubHTTPClient) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"status":"success"}`))),
	}, nil
}

func processTasks(documents TmcDocumentsResponse, baseURL string, client HTTPClient) ([]TaskResult, error) {
	var wg sync.WaitGroup
	results := make([]TaskResult, len(documents))
	errChan := make(chan error, len(documents))

	for i, doc := range documents {
		wg.Add(1)
		go func(idx int, doc TmcDocument) {
			defer wg.Done()
			err := processSingleTask(doc, baseURL, client)
			results[idx] = TaskResult{
				DocumentGUID: doc.TmcDocumentGUID,
				Success:      err == nil,
				Error:        err,
			}
			if err != nil {
				errChan <- err
			}
		}(i, doc)
	}

	wg.Wait()
	close(errChan)

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}
	if len(errors) > 0 {
		return results, fmt.Errorf("errors occurred during processing: %v", errors)
	}

	return results, nil
}

// Модифицированная processSingleTask
func processSingleTask(doc TmcDocument, baseURL string, client HTTPClient) error {
	requestBody, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("json.Marshal failed for document %s: %w", doc.TmcDocumentGUID, err)
	}

	requestURL := fmt.Sprintf("%s/%s", baseURL, doc.TmcDocumentGUID)
	resp, err := client.Post(requestURL, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("HTTP request failed for document %s: %w", doc.TmcDocumentGUID, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code for document %s: %d", doc.TmcDocumentGUID, resp.StatusCode)
	}

	return nil
}

func main() {
	// Данные для обработки
	jsonData := []byte(`[{
        "TmcDocumentGUID": "de3ed5f5-5295-11f0-93aa-be3af2b6059f",
        "TmcDocumentNumber": "000006376",
        "CreateTime": 1750957213,
        "Status": "Создано",
        "AcsessTime": 0,
        "ServiceGUID": "badb828d-ddbe-11ed-bba5-6cb31109ce4a",
        "EmployerGUID": "df8a96c3-0b9e-11ef-bbbb-6cb31109ce4a",
        "DriverGUID": "9dd2f73c-57b4-11ea-bb95-98f2b3136b67",
        "TmcList": [
            {
                "TmcGUID": "351b0a65-ea78-11e9-a836-9cdc71b5f11b",
                "TmcName": "Литол-24 ",
                "UnitOfMeasurement": "кг",
                "Count": 1,
                "Enable": false
            }
        ]
    },
	{
        "TmcDocumentGUID": "de3ed5f5-5295-11f0-93aa-be3af2b6059r",
        "TmcDocumentNumber": "000006376",
        "CreateTime": 1750957213,
        "Status": "Создано",
        "AcsessTime": 0,
        "ServiceGUID": "badb828d-ddbe-11ed-bba5-6cb31109ce4a",
        "EmployerGUID": "df8a96c3-0b9e-11ef-bbbb-6cb31109ce4a",
        "DriverGUID": "9dd2f73c-57b4-11ea-bb95-98f2b3136b67",
        "TmcList": [
            {
                "TmcGUID": "351b0a65-ea78-11e9-a836-9cdc71b5f11b",
                "TmcName": "Литол-24 ",
                "UnitOfMeasurement": "кг",
                "Count": 1,
                "Enable": false
            }
        ]
    }
	]`)

	var documents TmcDocumentsResponse
	if err := json.Unmarshal(jsonData, &documents); err != nil {
		log.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	// заглушка
	client := &StubHTTPClient{}

	results, err := processTasks(documents, "http://stub.api", client)
	if err != nil {
		log.Printf("processTasks failed: %v", err)
		return
	}

	if len(results) != 2 {
		log.Printf("Expected 2 result, got %d", len(results))
		return
	}
	if !results[0].Success {
		log.Printf("Expected successful processing for document %s, got error: %v", results[0].DocumentGUID, results[0].Error)
		return
	}

	log.Printf("All documents processed successfully:")
	for _, result := range results {
		log.Printf("Document %s: Success=%v, Error=%v", result.DocumentGUID, result.Success, result.Error)
	}
}
