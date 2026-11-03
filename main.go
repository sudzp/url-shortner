package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

// URLShortener handles thread-safe data mapping using a sync.Mutex
type URLShortener struct {
	mu   sync.Mutex
	urls map[string]string
}

// ShortenRequest defines the expected JSON input payload
type ShortenRequest struct {
	LongURL string `json:"url"`
}

// ShortenResponse defines the JSON response payload
type ShortenResponse struct {
	ShortURL string `json:"short_url"`
}

// NewURLShortener initializes the custom storage map
func NewURLShortener() *URLShortener {
	return &URLShortener{
		urls: make(map[string]string),
	}
}

// generateShortKey creates a random 6-character string for the short URL path
func generateShortKey() string {
	bytes := make([]byte, 3)
	if _, err := rand.Read(bytes); err != nil {
		return "short"
	}
	return hex.EncodeToString(bytes)
}

// HandleShorten processes POST requests to create a shortened URL
func (s *URLShortener) HandleShorten(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ShortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.LongURL == "" {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	key := generateShortKey()
	s.urls[key] = req.LongURL
	s.mu.Unlock()

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	shortenedURL := fmt.Sprintf("%s://%s/%s", scheme, r.Host, key)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ShortenResponse{ShortURL: shortenedURL})
}

// HandleRedirect processes GET requests and redirects the short key to the original URL
func (s *URLShortener) HandleRedirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the short key from the path (skipping the leading '/')
	key := r.URL.Path[1:]

	s.mu.Lock()
	longURL, exists := s.urls[key]
	s.mu.Unlock()

	if !exists {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusFound)
}

func main() {
	shortener := NewURLShortener()

	// Route definitions
	http.HandleFunc("/shorten", shortener.HandleShorten)
	http.HandleFunc("/", shortener.HandleRedirect)

	port := ":8080"
	log.Printf("Server starting on %s...\n", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
