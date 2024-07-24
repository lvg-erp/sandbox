package alg

import (
	"fmt"
	"net/http"
)

func FetchAPI(url string, results chan<- string) {
	resp, err := http.Get(url)
	if err != nil {
		results <- fmt.Sprintf("%s - not ok", url)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			results <- fmt.Sprintf("%s - ok", url)
		} else {
			results <- fmt.Sprintf("%s - not ok", url)
		}
	}
}
