package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	bakeryHandler "bakery-scheduler/internal/delivery/http/bakery"
)

func newTestServer() http.Handler {
	// handler dengan service nil cukup untuk tes ini: request ditolak middleware sebelum sampai ke service
	return (&Server{Bakery: bakeryHandler.New(nil), JWTSecret: []byte("x"), AllowedOrigins: []string{"http://localhost:3000"}}).Routes()
}

func TestCORSPreflight(t *testing.T) {
	h := newTestServer()
	req := httptest.NewRequest(http.MethodOptions, "/api/orders", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent || rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Fatalf("preflight: %d %v", rec.Code, rec.Header())
	}
}

func TestCORSRejectsUnknownOriginAndKeepsAuth(t *testing.T) {
	h := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Origin", "https://jahat.example")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("asal tak dikenal tidak boleh mendapat header CORS")
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("tanpa token harus 401, dapat %d", rec.Code)
	}
	// asal sah tetap harus melewati autentikasi, dan error 401 harus terbaca oleh browser
	req = httptest.NewRequest(http.MethodGet, "/api/orders", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rec = httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized || rec.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Errorf("401 + header CORS diharapkan: %d %v", rec.Code, rec.Header())
	}
}
