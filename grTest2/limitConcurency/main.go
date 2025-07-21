package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type Url struct {
	UrlID string `json:"UrlID"`
	Url   string `json:"Url"`
}

type Result struct {
	UrlID      string
	StatusCode int
	Success    bool
	Error      error
}

type APIClient interface {
	SendRequest(ctx context.Context, url string) (*http.Response, error)
}

type RestClient struct {
	client  *http.Client
	baseUrl string
}

func NewRestClient() *RestClient {
	return &RestClient{
		client: &http.Client{
			Timeout: 3 * time.Second,
		},
		baseUrl: "",
	}
}

func main() {
	// неправильные  []string{"http://example.com", "http://example.org", "http://example_reg.com", "http://example_pan.com"}
	urls := []string{"http://example.com", "http://example.org", "http://example.net", "http://example.edu"}
	client := NewRestClient()
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	result := processRequest(context.Background(), urls, 3, client)
	fmt.Println(result)
}

func processRequest(ctx context.Context, urls []string, working int, client APIClient) []Result {
	if working <= 0 {
		working = 1
	}
	var sCode []Result
	semaphore := make(chan struct{}, working)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(idx int, u string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			//maxAttempts := 3
			//for attempt := 1; attempt <= maxAttempts; attempt++ {
			//	// Проверяем, не отменен ли контекст
			//	if ctx.Err() != nil {
			//		mu.Lock()
			//		sCode = append(sCode, Result{
			//			UrlID:      strconv.Itoa(idx),
			//			Success:    false,
			//			StatusCode: -1,
			//			Error:      ctx.Err(),
			//		})
			//		mu.Unlock()
			//		return
			//	}

			resp, err := client.SendRequest(ctx, u)
			if err != nil {
				//log.Printf("Attempt %d for %s failed: %v", attempt, u, err)
				//if attempt < maxAttempts {
				//	time.Sleep(time.Second * time.Duration(attempt)) // Экспоненциальная задержка
				//	continue
				//}
				//
				mu.Lock()
				sCode = append(sCode, Result{
					UrlID:      strconv.Itoa(idx),
					Success:    false,
					StatusCode: -1,
					Error:      err,
				})
				mu.Unlock()
				return
			}

			// Успешный запрос
			defer resp.Body.Close()
			mu.Lock()
			sCode = append(sCode, Result{
				UrlID:      strconv.Itoa(idx),
				Success:    true,
				StatusCode: resp.StatusCode,
				Error:      nil,
			})
			mu.Unlock()
			return
			//}
		}(i, url)
	}

	wg.Wait()
	return sCode
}

func (c *RestClient) SendRequest(ctx context.Context, url string) (*http.Response, error) {

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
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
