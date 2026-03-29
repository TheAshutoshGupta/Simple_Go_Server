package main

import (
	"fmt"
	"net/http"
	"time"
)

type Bucket struct {
	tokens         float64
	lastRefillTime time.Time
}

const AllowedTokenPerSecond = 10 // max requests per second
const capacity = 10          // max tokens in the bucket

var rateLimits = make(map[string]Bucket)

func checkRateLimit(userID string) bool {
	now := time.Now()
	bucket, exists := rateLimits[userID]
	if !exists {
		bucket = Bucket{tokens: capacity - 1, lastRefillTime: now}
	}

	elapsedTime := now.Sub(bucket.lastRefillTime).Seconds()
	bucket.tokens += elapsedTime * float64(AllowedTokenPerSecond)
	if bucket.tokens > float64(capacity) {
		bucket.tokens = float64(capacity)
	}

	if bucket.tokens < 1 {
		rateLimits[userID] = bucket
		return false
	}
	bucket.lastRefillTime = now
	bucket.tokens -= 1 // consume a token for the request
	rateLimits[userID] = bucket
	return true
}

// handler function for GET /user
func testUser(w http.ResponseWriter, r *http.Request) {
	// only allow GET requests

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID := r.URL.Query().Get("id")
	isAllowed := checkRateLimit(userID)
	if !isAllowed {
		http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
		return
	}
	http.ServeFile(w, r, "index.html")
}

func main() {
	http.HandleFunc("/test", testUser)

	fmt.Println("server running on port 8080")
	http.ListenAndServe(":8080", nil)
}
