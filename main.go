// URL Shortener with Concurrent Click Analytics
//
// Demonstrates Go's core strength: goroutines and channels for
// concurrent processing. Every click redirect is instant - the
// actual database write for analytics happens asynchronously in a
// background worker, consuming from a buffered channel. This is a
// standard real-world Go pattern for decoupling a fast-path (the
// redirect) from a slow-path (analytics writes) without blocking
// the user-facing request.

package main

import (
"crypto/rand"
"database/sql"
"encoding/json"
"fmt"
"log"
"math/big"
"net/http"
"os"

_ "github.com/lib/pq"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var db *sql.DB
var clickEvents chan string

type ShortenRequest struct {
URL string `json:"url"`
}

type ShortenResponse struct {
ShortCode string `json:"short_code"`
ShortURL  string `json:"short_url"`
}

// generateShortCode creates a random 6-character alphanumeric code.
func generateShortCode() (string, error) {
code := make([]byte, 6)
for i := range code {
num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
if err != nil {
return "", err
}
code[i] = charset[num.Int64()]
}
return string(code), nil
}

// clickWorker runs in the background, consuming click events from the
// channel and writing them to the database - decoupled from the
// user-facing redirect, which never waits on this.
func clickWorker() {
for shortCode := range clickEvents {
_, err := db.Exec(
"INSERT INTO click_events (short_code) VALUES ($1)",
shortCode,
)
if err != nil {
log.Printf("Failed to record click for %s: %v", shortCode, err)
}
}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

func shortenHandler(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodPost {
http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
return
}

var req ShortenRequest
if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.URL == "" {
http.Error(w, `{"error":"Invalid request body"}`, http.StatusUnprocessableEntity)
return
}

shortCode, err := generateShortCode()
if err != nil {
http.Error(w, `{"error":"Failed to generate short code"}`, http.StatusInternalServerError)
return
}

_, err = db.Exec(
"INSERT INTO short_urls (short_code, original_url) VALUES ($1, $2)",
shortCode, req.URL,
)
if err != nil {
http.Error(w, `{"error":"Failed to save URL"}`, http.StatusInternalServerError)
return
}

resp := ShortenResponse{
ShortCode: shortCode,
ShortURL:  fmt.Sprintf("http://localhost:8080/%s", shortCode),
}
w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(resp)
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
shortCode := r.URL.Path[1:]

var originalURL string
err := db.QueryRow(
"SELECT original_url FROM short_urls WHERE short_code = $1",
shortCode,
).Scan(&originalURL)

if err == sql.ErrNoRows {
http.Error(w, `{"error":"Short URL not found"}`, http.StatusNotFound)
return
}
if err != nil {
http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
return
}

// Non-blocking send: the redirect happens immediately, the click
// is recorded asynchronously by the background worker.
select {
case clickEvents <- shortCode:
default:
log.Printf("Click event channel full, dropping event for %s", shortCode)
}

http.Redirect(w, r, originalURL, http.StatusFound)
}

func statsHandler(w http.ResponseWriter, r *http.Request) {
shortCode := r.URL.Query().Get("code")
if shortCode == "" {
http.Error(w, `{"error":"code query parameter required"}`, http.StatusUnprocessableEntity)
return
}

var clickCount int
err := db.QueryRow(
"SELECT COUNT(*) FROM click_events WHERE short_code = $1",
shortCode,
).Scan(&clickCount)

if err != nil {
http.Error(w, `{"error":"Database error"}`, http.StatusInternalServerError)
return
}

w.Header().Set("Content-Type", "application/json")
json.NewEncoder(w).Encode(map[string]int{"click_count": clickCount})
}

func main() {
var err error
dbUser := os.Getenv("PGUSER")
	if dbUser == "" {
		dbUser = "postgres"
	}
	db, err = sql.Open("postgres", fmt.Sprintf("dbname=mydb sslmode=disable user=%s", dbUser))
if err != nil {
log.Fatal(err)
}
defer db.Close()

if err := db.Ping(); err != nil {
log.Fatal("Cannot connect to database: ", err)
}

clickEvents = make(chan string, 100)
go clickWorker()

http.HandleFunc("/health", healthHandler)
http.HandleFunc("/shorten", shortenHandler)
http.HandleFunc("/stats", statsHandler)
http.HandleFunc("/", redirectHandler)

fmt.Println("URL Shortener running on http://localhost:8080")
log.Fatal(http.ListenAndServe(":8080", nil))
}
