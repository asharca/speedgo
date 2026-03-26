package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateTestData(t *testing.T) {
	data := generateTestData(1024)
	if len(data) != 1024 {
		t.Errorf("expected 1024 bytes, got %d", len(data))
	}
}

func TestGenerateTestData_Large(t *testing.T) {
	size := 1 * 1024 * 1024
	data := generateTestData(size)
	if len(data) != size {
		t.Errorf("expected %d bytes, got %d", size, len(data))
	}
}

func TestGetUploadEndpoint_Global(t *testing.T) {
	ep := getUploadEndpoint(RegionGlobal)
	if ep == "" {
		t.Error("expected non-empty endpoint for global")
	}
}

func TestGetUploadEndpoint_CN(t *testing.T) {
	ep := getUploadEndpoint(RegionCN)
	if ep == "" {
		t.Error("expected non-empty endpoint for cn")
	}
}

func TestGetUploadEndpoint_Unknown(t *testing.T) {
	ep := getUploadEndpoint(Region("mars"))
	if ep != "https://speed.cloudflare.com/__up" {
		t.Errorf("expected fallback endpoint, got %s", ep)
	}
}

func TestUploadChunk(t *testing.T) {
	var received int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = len(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	data := generateTestData(512)
	bytesChan := make(chan int64, 10)

	client := &http.Client{Timeout: 5 * time.Second}
	err := uploadChunk(context.Background(), client, srv.URL, data, bytesChan)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received != 512 {
		t.Errorf("server received %d bytes, want 512", received)
	}

	select {
	case b := <-bytesChan:
		if b != 512 {
			t.Errorf("reported %d bytes, want 512", b)
		}
	default:
		t.Error("expected bytes report in channel")
	}
}

func TestUploadChunk_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer srv.Close()

	data := generateTestData(256)
	bytesChan := make(chan int64, 10)
	client := &http.Client{Timeout: 5 * time.Second}

	err := uploadChunk(context.Background(), client, srv.URL, data, bytesChan)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestUploadChunk_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	data := generateTestData(256)
	bytesChan := make(chan int64, 10)
	client := &http.Client{Timeout: 5 * time.Second}

	err := uploadChunk(ctx, client, srv.URL, data, bytesChan)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestCountingReader(t *testing.T) {
	data := []byte("hello world")
	cr := &countingReader{
		reader: io.NopCloser(nil), // will be overridden
	}
	// Test with actual bytes reader
	cr2 := &countingReader{
		reader: io.LimitReader(
			&infiniteReader{},
			100,
		),
	}
	buf := make([]byte, 50)
	n, _ := cr2.Read(buf)
	if n != 50 {
		t.Errorf("expected 50 bytes read, got %d", n)
	}
	if cr2.count != 50 {
		t.Errorf("count = %d, want 50", cr2.count)
	}
	_ = data
	_ = cr
}

type infiniteReader struct{}

func (r *infiniteReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 'x'
	}
	return len(p), nil
}

func TestParseUploadConfig(t *testing.T) {
	config, err := parseUploadConfig([]string{"-duration=20", "-concurrency=2", "-region=global"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Duration != 20*time.Second {
		t.Errorf("Duration = %v, want 20s", config.Duration)
	}
	if config.Concurrency != 2 {
		t.Errorf("Concurrency = %d, want 2", config.Concurrency)
	}
	if config.Region != RegionGlobal {
		t.Errorf("Region = %s, want global", config.Region)
	}
}

func TestParseUploadConfig_RegionCN(t *testing.T) {
	config, err := parseUploadConfig([]string{"-duration=5", "-concurrency=1", "-region=cn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionCN {
		t.Errorf("Region = %s, want cn", config.Region)
	}
	if config.Duration != 5*time.Second {
		t.Errorf("Duration = %v, want 5s", config.Duration)
	}
}

func TestMeasureUploadSpeed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	config := &UploadConfig{
		Duration:    2 * time.Second,
		Concurrency: 1,
		Region:      RegionGlobal,
	}

	stats := measureUploadSpeed(context.Background(), config, srv.URL)
	if stats.BytesSent == 0 {
		t.Error("expected some bytes sent")
	}
	if stats.Speed <= 0 {
		t.Error("expected positive speed")
	}
}
