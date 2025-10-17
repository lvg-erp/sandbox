package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// фейковый клиент
type RestClient struct {
	client  *http.Client
	baseUrl string
}

// Структура хранения результата
type Result struct {
	UrlID      string
	StatusCode int
	Success    bool
	Error      error
	Target     bool
}

// инетерфейс обработки
type APIClient interface {
	SendRequest(ctx context.Context, url string) (*http.Response, error)
}

func NewRestClient() *RestClient {
	return &RestClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseUrl: "",
	}
}

func main() {
	urls := []string{"http://example.com", "http://example.org", "http://tritri.net", "http://example.edu"}
	//client := NewRestClient()
	//ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()
	result, err := ParallelSearch(context.Background(), urls, "exampl")
	if err != nil {
		fmt.Errorf("this err %d", err)
	}
	fmt.Println(result)
}

func ParallelSearch(ctx context.Context, urls []string, target string) ([]string, error) {
	var (
		result []string
		mu     sync.Mutex
		wg     sync.WaitGroup
		err    error
	)
	client := &http.Client{Timeout: 5 * time.Second}
	semaphore := make(chan struct{}, len(urls))

	for _, url := range urls {
		wg.Add(1)
		semaphore <- struct{}{}
		go func(u string) {
			defer wg.Done()
			defer func() { <-semaphore }()

			req, errReq := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
			if errReq != nil {
				mu.Lock()
				err = errReq
				mu.Unlock()
				return
			}

			resp, errResp := client.Do(req)
			if errResp != nil {
				mu.Lock()
				err = errResp
				mu.Unlock()
				return
			}
			defer resp.Body.Close()

			body, errBody := io.ReadAll(resp.Body)
			if errBody != nil {
				mu.Lock()
				err = errBody
				mu.Unlock()
				return
			}

			if strings.Contains(string(body), target) {
				mu.Lock()
				result = append(result, u)
				mu.Unlock()
			}
		}(url)
	}

	wg.Wait()
	return result, err
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

func searchInResponse(ctx context.Context, url, target string) (bool, error) {
	// Создаем HTTP-клиент
	client := &http.Client{Timeout: 5 * time.Second}

	// Создаем запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, err
	}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close() // Обязательно закрываем тело ответа

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}

	// Ищем подстроку в теле ответа
	found := strings.Contains(string(body), target)
	return found, nil
}

func findSubstring(str string, subStr string) bool {
	r := []rune(str)
	sr := []rune(subStr)
	if len(sr) == 0 {
		return false
	}
str:
	for i, ru := range r {
		if ru == sr[0] {
			for j, x := range sr[1:] {
				if r[i+j+1] != x {
					continue str
				}
			}
			return true
		}
	}
	return false
}

//func ParallelSearch(ctx context.Context, urls []string, working int, client APIClient, target string) []Result {
//
//	if working == 0 {
//		working = 1
//	}
//
//	var (
//		result []Result
//		wg     sync.WaitGroup
//		mu     sync.Mutex
//	)
//	semaphore := make(chan struct{}, working)
//
//	for i, url := range urls {
//		wg.Add(1)
//		semaphore <- struct{}{}
//		go func(idx int, u string) {
//			defer wg.Done()
//			defer func() { <-semaphore }()
//
//			resp, err := client.SendRequest(ctx, url)
//			// проверяем ответ на запрос, если нет ответа(404, 400....)
//			if err != nil {
//				mu.Lock()
//				result = append(result, Result{
//					UrlID:      url,
//					Success:    false,
//					Error:      err,
//					StatusCode: resp.StatusCode,
//					Target:     false,
//				})
//				mu.Unlock()
//				return
//			}
//
//			//запрос корректен - ищем трагет
//			trg := findSubstring(url, target)
//			//успешный запрос
//			defer resp.Body.Close()
//			mu.Lock()
//			result = append(result, Result{
//				UrlID:      url,
//				Success:    true,
//				StatusCode: resp.StatusCode,
//				Error:      nil,
//				Target:     trg,
//			})
//			mu.Unlock()
//			return
//
//		}(i, url)
//	}
//
//	wg.Wait()
//	return result
//
//}
