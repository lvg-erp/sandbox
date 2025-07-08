package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

const (
	maxWorkers = 8 // максимальное количество горутин
)

type Order struct {
	OrderID string  `json:"orderId"`
	Amount  float64 `json:"amount"`
}

type Result struct {
	OrderID string
	Success bool
	Error   error
}

type APIClient interface {
	SendOrder(ctx context.Context, order Order) (*http.Response, error)
}

type RestClient struct {
	client  *http.Client
	baseUrl string
}

func NewRestClient(baseUrl string) *RestClient {
	return &RestClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseUrl: baseUrl,
	}
}

func main() {
	jsonData := `
    [
      {
        "OrderID": "doc-1-5295-11f0-93aa-be3af2b6059f",
        "Amount": 20000.50
      },
      {
        "OrderID": "doc-1-5295-11f0-93aa-be3af2b605ss",
        "Amount": 21500.50
      },
      {
        "OrderID": "doc-1-5295-11f0-93aa-be3af2b605rr",
        "Amount": 25000.30
      },
	  {
        "OrderID": "doc-1-5295-11f0-93aa-be3af2b605mm",
        "Amount": 20000.50
      },
      {
        "OrderID": "doc-1-5295-11f0-93aa-be3af2b605gg",
        "Amount": 21500.50
      }	
    ]`

	var orders []Order
	if err := json.Unmarshal([]byte(jsonData), &orders); err != nil {
		log.Fatalf("Failed to unmarshal jsonData: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRestClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := ProcessOrders(ctx, orders, client)
	if err != nil {
		log.Printf("ProcessOrders failed: %v", err)
		return
	}

	for _, result := range results {
		if result.Success {
			log.Printf("Order %s processed successfully", result.OrderID)
		} else {
			log.Printf("Order %s failed: %v", result.OrderID, result.Error)
		}
	}
}

func ProcessOrders(ctx context.Context, orders []Order, client APIClient) ([]Result, error) {
	results := make([]Result, len(orders))
	//errChan:=(chan error, len(orders))
	semaphore := make(chan struct{}, maxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for i, order := range orders {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(i int, order Order) {
			defer wg.Done()
			defer func() { <-semaphore }()
			log.Printf("Starting processing send order %s at %v", order.OrderID, time.Now())

			if err := ctx.Err(); err != nil {
				mu.Lock()
				results[i] = Result{OrderID: order.OrderID, Success: false, Error: err}
				log.Printf("Finished processing order %s with context error: %v", order.OrderID, err)
				mu.Unlock()
				return
			}

			resp, err := client.SendOrder(ctx, order)
			if err != nil {
				mu.Lock()
				results[i] = Result{OrderID: order.OrderID, Success: false, Error: err}
				mu.Unlock()
				log.Printf("Finished processing order %s with error: %v", order.OrderID, err)
				return
			}

			defer resp.Body.Close()
			mu.Lock()
			if resp.StatusCode == http.StatusOK {
				results[i] = Result{OrderID: order.OrderID, Success: true, Error: nil}
			} else {
				results[i] = Result{OrderID: order.OrderID, Success: false, Error: err}
			}
			mu.Unlock()
		}(i, order)

	}
	wg.Wait()
	for _, res := range results {
		if res.Error != nil {
			return results, fmt.Errorf("one or more orders failed: %v", res.Error)
		}
	}

	return results, nil
}

func (c *RestClient) SendOrder(ctx context.Context, order Order) (*http.Response, error) {

	body, err := json.Marshal(order)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil

}
