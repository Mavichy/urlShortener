package httpserver

import (
	"bytes"
	"encoding/json"
	"github.com/Mavichy/urlShortener/internal/service"
	"github.com/Mavichy/urlShortener/internal/shortcode"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	memorystore "github.com/Mavichy/urlShortener/internal/storage/memory"
)

func TestCreateAndResolveHandlers(t *testing.T) {
	svc := service.New(memorystore.New(), shortcode.NewGenerator())

	srv := New(ServerConfig{
		Addr:           ":8080",
		BaseURL:        "http://localhost:8080",
		RequestTimeout: 2 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    2 * time.Second,
	}, svc)

	createReq := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/links",
		bytes.NewBufferString(`{"url":"https://example.com/very/long"}`),
	)
	createReq.Header.Set("Content-Type", "application/json")

	createRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(createRec, createReq)

	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	var createResp map[string]string
	if err := json.NewDecoder(createRec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	code := createResp["code"]
	if code == "" {
		t.Fatal("expected non-empty code")
	}

	resolveReq := httptest.NewRequest(http.MethodGet, "/"+code, nil)
	resolveRec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(resolveRec, resolveReq)

	if resolveRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, resolveRec.Code)
	}

	var resolveResp map[string]string
	if err := json.NewDecoder(resolveRec.Body).Decode(&resolveResp); err != nil {
		t.Fatalf("decode resolve response: %v", err)
	}

	if resolveResp["original_url"] != "https://example.com/very/long" {
		t.Fatalf("unexpected original_url: %q", resolveResp["original_url"])
	}
}

func TestCreateHandler_InvalidJSON(t *testing.T) {
	svc := service.New(memorystore.New(), shortcode.NewGenerator())

	srv := New(ServerConfig{
		Addr:           ":8080",
		BaseURL:        "http://localhost:8080",
		RequestTimeout: 2 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    2 * time.Second,
	}, svc)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/links",
		bytes.NewBufferString(`{"url":123}`),
	)
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestResolveHandler_InvalidCode(t *testing.T) {
	svc := service.New(memorystore.New(), shortcode.NewGenerator())

	srv := New(ServerConfig{
		Addr:           ":8080",
		BaseURL:        "http://localhost:8080",
		RequestTimeout: 2 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    2 * time.Second,
	}, svc)

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestResolveHandler_NotFound(t *testing.T) {
	svc := service.New(memorystore.New(), shortcode.NewGenerator())

	srv := New(ServerConfig{
		Addr:           ":8080",
		BaseURL:        "http://localhost:8080",
		RequestTimeout: 2 * time.Second,
		ReadTimeout:    2 * time.Second,
		WriteTimeout:   2 * time.Second,
		IdleTimeout:    2 * time.Second,
	}, svc)

	req := httptest.NewRequest(http.MethodGet, "/AbCd1234_x", nil)
	rec := httptest.NewRecorder()

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
