package core

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProbeServers_SortsByLatency(t *testing.T) {
	// Create two test servers: one fast, one slow
	fast := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer fast.Close()

	slow := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer slow.Close()

	servers := []TestServer{
		{Name: "Slow", URL: slow.URL + "/test", Region: RegionGlobal, Provider: "test"},
		{Name: "Fast", URL: fast.URL + "/test", Region: RegionGlobal, Provider: "test"},
	}

	results := ProbeServers(context.Background(), servers, 5*time.Second)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].Server.Name != "Fast" {
		t.Errorf("expected Fast server first, got %s", results[0].Server.Name)
	}
	if results[0].Error != nil {
		t.Errorf("fast server should not have error: %v", results[0].Error)
	}
}

func TestProbeServers_UnreachableServer(t *testing.T) {
	good := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer good.Close()

	servers := []TestServer{
		{Name: "Bad", URL: "http://192.0.2.1:1/nope", Region: RegionGlobal, Provider: "test"},
		{Name: "Good", URL: good.URL + "/test", Region: RegionGlobal, Provider: "test"},
	}

	results := ProbeServers(context.Background(), servers, 2*time.Second)

	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	// Good server should be first (bad has error, goes to end)
	if results[0].Server.Name != "Good" {
		t.Errorf("expected Good server first, got %s", results[0].Server.Name)
	}
	if results[1].Error == nil {
		t.Error("expected error for unreachable server")
	}
}

func TestProbeServers_Empty(t *testing.T) {
	results := ProbeServers(context.Background(), nil, time.Second)
	if len(results) != 0 {
		t.Errorf("expected 0 results for nil input, got %d", len(results))
	}
}

func TestProbeServers_ContextCancelled(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	servers := []TestServer{
		{Name: "Slow", URL: srv.URL, Region: RegionGlobal, Provider: "test"},
	}

	results := ProbeServers(ctx, servers, 100*time.Millisecond)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Error == nil {
		t.Error("expected error for timed-out probe")
	}
}

func TestSelectBestServers_ReturnsN(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv1.Close()
	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv2.Close()
	srv3 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv3.Close()

	servers := []TestServer{
		{Name: "S1", URL: srv1.URL, Region: RegionGlobal, Provider: "test"},
		{Name: "S2", URL: srv2.URL, Region: RegionGlobal, Provider: "test"},
		{Name: "S3", URL: srv3.URL, Region: RegionGlobal, Provider: "test"},
	}

	best := SelectBestServers(context.Background(), servers, 2)
	if len(best) > 2 {
		t.Errorf("expected at most 2 servers, got %d", len(best))
	}
	if len(best) == 0 {
		t.Error("expected at least 1 server")
	}
}

func TestSelectBestServers_Empty(t *testing.T) {
	best := SelectBestServers(context.Background(), nil, 3)
	if best != nil {
		t.Errorf("expected nil for empty input, got %v", best)
	}
}

func TestSelectBestServers_AllFail(t *testing.T) {
	servers := []TestServer{
		{Name: "Bad1", URL: "http://192.0.2.1:1/nope", Region: RegionGlobal, Provider: "test"},
		{Name: "Bad2", URL: "http://192.0.2.2:1/nope", Region: RegionGlobal, Provider: "test"},
	}

	best := SelectBestServers(context.Background(), servers, 2)
	// Should fallback to all servers
	if len(best) != len(servers) {
		t.Errorf("expected fallback to all %d servers, got %d", len(servers), len(best))
	}
}

func TestProbeLatency_ValidServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "HEAD" {
			t.Errorf("expected HEAD request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	latency, err := probeLatency(context.Background(), srv.URL+"/file.dat", 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latency <= 0 {
		t.Errorf("expected positive latency, got %v", latency)
	}
}

func TestProbeLatency_InvalidURL(t *testing.T) {
	_, err := probeLatency(context.Background(), "://invalid", time.Second)
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}
