package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

var SitesPingedCount atomic.Int32

// RateLimiter manages global rate limiting
type RateLimiter struct {
	mu       sync.Mutex
	lastPing time.Time
}

var limiter = &RateLimiter{}

func (rl *RateLimiter) isAllowed() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	if now.Sub(rl.lastPing) < 10*time.Second {
		return false
	}

	rl.lastPing = now
	return true
}

func main() {
	http.HandleFunc("/ping-site", handlePingSite)
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/ping-count", handlePingCount)
	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	log.Println("Index page requested")
	http.ServeFile(w, r, "web/index.html")
}

func handlePingCount(w http.ResponseWriter, r *http.Request) {
	log.Printf("Ping count requested %d", SitesPingedCount.Load())
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"pingCount": int(SitesPingedCount.Load())})
}

func handlePingSite(w http.ResponseWriter, r *http.Request) {
	// Check global rate limit
	if !limiter.isAllowed() {
		log.Printf("Rate limit exceeded for ALL")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Rate limit exceeded. Please wait 10 seconds between pinging websites. This is to prevent abuse and not overload my server. This rate limit is applied to all users. ;) Thanks for your patience!",
		})
		return
	}

	url := r.URL.Query().Get("url")
	if url == "" {
		log.Printf("URL parameter is missing")
		http.Error(w, "URL parameter is missing", http.StatusBadRequest)
		return
	}

	// Add scheme if not present
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}

	// Fetch the webpage
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error fetching URL: %v", err)
		http.Error(w, "Error fetching URL", http.StatusBadRequest)
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		http.Error(w, "Error reading response", http.StatusInternalServerError)
		return
	}

	// Create response structure
	response := struct {
		HTML       string `json:"html"`
		StatusCode int    `json:"statusCode"`
	}{
		HTML:       string(body),
		StatusCode: resp.StatusCode,
	}
	SitesPingedCount.Add(1)
	log.Printf("Ping site requested. Total successful sites pinged: %d", SitesPingedCount.Load())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
