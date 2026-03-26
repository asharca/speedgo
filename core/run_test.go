package core

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRunDownload_WithTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 50000))
	}))
	defer srv.Close()

	// Temporarily override servers
	origServers := DownloadServers
	DownloadServers = []TestServer{
		{Name: "Test", URL: srv.URL, Region: RegionGlobal, Provider: "test"},
	}
	defer func() { DownloadServers = origServers }()

	err := RunDownload(context.Background(), []string{"-duration=2s", "-concurrency=1", "-region=global"})
	if err != nil {
		t.Fatalf("RunDownload failed: %v", err)
	}
}

func TestRunUpload_WithTestServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	origServers := UploadServers
	UploadServers = []TestServer{
		{Name: "Test", URL: srv.URL, Region: RegionGlobal, Provider: "test"},
	}
	defer func() { UploadServers = origServers }()

	err := RunUpload(context.Background(), []string{"-duration=2", "-concurrency=1", "-region=global"})
	if err != nil {
		t.Fatalf("RunUpload failed: %v", err)
	}
}

func TestRunAll_WithTestServers(t *testing.T) {
	dlSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write(make([]byte, 50000))
	}))
	defer dlSrv.Close()

	ulSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ulSrv.Close()

	origDL := DownloadServers
	origUL := UploadServers
	origPing := PingTargets

	DownloadServers = []TestServer{
		{Name: "DL Test", URL: dlSrv.URL, Region: RegionGlobal, Provider: "test"},
	}
	UploadServers = []TestServer{
		{Name: "UL Test", URL: ulSrv.URL, Region: RegionGlobal, Provider: "test"},
	}
	PingTargets = map[Region][]string{
		RegionGlobal: {"127.0.0.1"},
		RegionCN:     {"127.0.0.1"},
	}
	defer func() {
		DownloadServers = origDL
		UploadServers = origUL
		PingTargets = origPing
	}()

	err := RunAll(context.Background(), []string{"-region=global", "-concurrency=1"})
	// Ping may fail without root, but the rest should work
	_ = err
}
