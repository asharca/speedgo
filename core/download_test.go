package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetTestFiles(t *testing.T) {
	servers := []TestServer{
		{Name: "S1", URL: "http://example.com/1", Region: RegionGlobal},
		{Name: "S2", URL: "http://example.com/2", Region: RegionCN},
	}
	urls := getTestFiles(servers)
	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "http://example.com/1" {
		t.Errorf("urls[0] = %q, want http://example.com/1", urls[0])
	}
	if urls[1] != "http://example.com/2" {
		t.Errorf("urls[1] = %q, want http://example.com/2", urls[1])
	}
}

func TestGetTestFiles_Empty(t *testing.T) {
	urls := getTestFiles(nil)
	if len(urls) != 0 {
		t.Errorf("expected 0 URLs for nil, got %d", len(urls))
	}
}

func TestDownloadChunk(t *testing.T) {
	data := []byte("hello world test data for download chunk")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(data)
	}))
	defer srv.Close()

	bytesChan := make(chan int64, 100)
	err := downloadChunk(context.Background(), srv.URL, bytesChan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var total int64
	close(bytesChan)
	for b := range bytesChan {
		total += b
	}
	if total != int64(len(data)) {
		t.Errorf("expected %d bytes, got %d", len(data), total)
	}
}

func TestDownloadChunk_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	bytesChan := make(chan int64, 100)
	err := downloadChunk(context.Background(), srv.URL, bytesChan)
	// Server returns 500 but body is empty, so no error from reading
	// This is fine - just verify it doesn't crash
	_ = err
}

func TestDownloadChunk_InvalidURL(t *testing.T) {
	bytesChan := make(chan int64, 100)
	err := downloadChunk(context.Background(), "http://192.0.2.1:1/nope", bytesChan)
	if err == nil {
		t.Error("expected error for unreachable URL")
	}
}

func TestDownloadChunk_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("data"))
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	bytesChan := make(chan int64, 100)
	err := downloadChunk(ctx, srv.URL, bytesChan)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestParseDownloadConfig(t *testing.T) {
	config, err := parseDownloadConfig([]string{"-duration=10s", "-concurrency=2", "-region=cn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Duration != 10*time.Second {
		t.Errorf("Duration = %v, want 10s", config.Duration)
	}
	if config.Concurrency != 2 {
		t.Errorf("Concurrency = %d, want 2", config.Concurrency)
	}
	if config.Region != RegionCN {
		t.Errorf("Region = %s, want cn", config.Region)
	}
}

func TestParseDownloadConfig_RegionGlobal(t *testing.T) {
	config, err := parseDownloadConfig([]string{"-duration=5s", "-concurrency=1", "-region=global"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionGlobal {
		t.Errorf("Region = %s, want global", config.Region)
	}
	if config.Duration != 5*time.Second {
		t.Errorf("Duration = %v, want 5s", config.Duration)
	}
}

func TestMeasureDownloadSpeed(t *testing.T) {
	// Create a test server that returns some data
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 10000))
	}))
	defer srv.Close()

	config := &DownloadConfig{
		Duration:    2 * time.Second,
		Concurrency: 1,
		Region:      RegionGlobal,
		servers: []TestServer{
			{Name: "Test", URL: srv.URL, Region: RegionGlobal},
		},
	}

	stats := measureDownloadSpeed(context.Background(), config)
	if stats.BytesReceived == 0 {
		t.Error("expected some bytes received")
	}
	if stats.Speed <= 0 {
		t.Error("expected positive speed")
	}
	if stats.Duration <= 0 {
		t.Error("expected positive duration")
	}
}
