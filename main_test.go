// Automated tests for the URL Shortener.
// Demonstrates Go's built-in testing package (no external framework
// needed) - table-driven tests and HTTP handler tests using
// net/http/httptest.

package main

import (
"encoding/json"
"net/http"
"net/http/httptest"
"strings"
"testing"
)

func TestGenerateShortCode(t *testing.T) {
code, err := generateShortCode()
if err != nil {
t.Fatalf("expected no error, got %v", err)
}
if len(code) != 6 {
t.Errorf("expected code length 6, got %d", len(code))
}
for _, c := range code {
if !strings.ContainsRune(charset, c) {
t.Errorf("code contains invalid character: %c", c)
}
}
}

func TestGenerateShortCodeIsRandom(t *testing.T) {
code1, _ := generateShortCode()
code2, _ := generateShortCode()
if code1 == code2 {
t.Error("expected two generated codes to differ (extremely unlikely collision, or generator is broken)")
}
}

func TestHealthHandler(t *testing.T) {
req := httptest.NewRequest(http.MethodGet, "/health", nil)
w := httptest.NewRecorder()

healthHandler(w, req)

if w.Code != http.StatusOK {
t.Errorf("expected status 200, got %d", w.Code)
}

var response map[string]string
if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
t.Fatalf("failed to decode response: %v", err)
}
if response["status"] != "healthy" {
t.Errorf("expected status 'healthy', got %s", response["status"])
}
}

func TestShortenHandlerRejectsWrongMethod(t *testing.T) {
req := httptest.NewRequest(http.MethodGet, "/shorten", nil)
w := httptest.NewRecorder()

shortenHandler(w, req)

if w.Code != http.StatusMethodNotAllowed {
t.Errorf("expected status 405, got %d", w.Code)
}
}

func TestShortenHandlerRejectsEmptyBody(t *testing.T) {
req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader("{}"))
w := httptest.NewRecorder()

shortenHandler(w, req)

if w.Code != http.StatusUnprocessableEntity {
t.Errorf("expected status 422, got %d", w.Code)
}
}

func TestStatsHandlerRequiresCodeParam(t *testing.T) {
req := httptest.NewRequest(http.MethodGet, "/stats", nil)
w := httptest.NewRecorder()

statsHandler(w, req)

if w.Code != http.StatusUnprocessableEntity {
t.Errorf("expected status 422, got %d", w.Code)
}
}
