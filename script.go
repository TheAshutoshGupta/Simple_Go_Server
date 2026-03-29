package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	url := "http://localhost:8080/test?id=123"

	concurrency := 50
	duration := 10 * time.Second

	// Proper HTTP client with connection reuse
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 100,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
	}

	var wg sync.WaitGroup
	var success, fail int
	var mu sync.Mutex

	end := time.Now().Add(duration)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				if time.Now().After(end) {
					return
				}

				resp, err := client.Get(url)
				if err != nil {
					continue
				}
				resp.Body.Close()

				mu.Lock()
				if resp.StatusCode == http.StatusOK {
					success++
				} else if resp.StatusCode == http.StatusTooManyRequests {
					fail++
				}
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	fmt.Println("==== RESULT ====")
	fmt.Println("Success:", success)
	fmt.Println("Rejected:", fail)
}