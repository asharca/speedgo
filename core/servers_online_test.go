//go:build online

package core

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"testing"
	"time"
)

// TestDownloadServersReachable verifies all download server URLs respond to GET.
// Run with: go test ./core/ -tags=online -run TestDownloadServersReachable -v
func TestDownloadServersReachable(t *testing.T) {
	ctx := context.Background()

	for _, s := range DownloadServers {
		t.Run(fmt.Sprintf("%s_%s", s.Region, s.Name), func(t *testing.T) {
			// Retry up to 2 times for transient network issues (TLS handshake, etc.)
			var lastErr error
			for attempt := 0; attempt < 2; attempt++ {
				reqCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
				req, err := http.NewRequestWithContext(reqCtx, "GET", s.URL, nil)
				if err != nil {
					cancel()
					t.Fatalf("creating request: %v", err)
				}
				SetUA(req)

				client := &http.Client{Timeout: 10 * time.Second}
				resp, err := client.Do(req)
				if err != nil {
					cancel()
					lastErr = err
					t.Logf("attempt %d failed: %v", attempt+1, err)
					time.Sleep(time.Second)
					continue
				}

				buf := make([]byte, 4096)
				n, _ := io.ReadAtLeast(resp.Body, buf, 1)
				resp.Body.Close()
				cancel()

				if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
					t.Fatalf("server %s returned status %d, want 200/206", s.Name, resp.StatusCode)
				}
				if n == 0 {
					t.Fatalf("server %s returned 0 bytes", s.Name)
				}
				t.Logf("OK: %s [%s] status=%d, read=%d bytes", s.Name, s.Provider, resp.StatusCode, n)
				return
			}
			t.Fatalf("server %s (%s) unreachable after retries: %v", s.Name, s.URL, lastErr)
		})
	}
}

// TestUploadServersReachable verifies all upload server URLs accept POST.
// Run with: go test ./core/ -tags=online -run TestUploadServersReachable -v
func TestUploadServersReachable(t *testing.T) {
	ctx := context.Background()

	for _, s := range UploadServers {
		t.Run(fmt.Sprintf("%s_%s", s.Region, s.Name), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			data := generateTestData(1024) // 1KB test payload
			req, err := http.NewRequestWithContext(ctx, "POST", s.URL,
				io.NopCloser(io.LimitReader(readerFromBytes(data), int64(len(data)))))
			if err != nil {
				t.Fatalf("creating request: %v", err)
			}
			SetUA(req)
			req.Header.Set("Content-Type", "application/octet-stream")

			client := &http.Client{Timeout: 15 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("server %s (%s) unreachable: %v", s.Name, s.URL, err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf("server %s returned status %d, want 200", s.Name, resp.StatusCode)
			}
			t.Logf("OK: %s [%s] status=%d", s.Name, s.Provider, resp.StatusCode)
		})
	}
}

// TestPingTargetsDNSResolvable verifies all ping targets can be DNS-resolved.
// Run with: go test ./core/ -tags=online -run TestPingTargetsDNSResolvable -v
func TestPingTargetsDNSResolvable(t *testing.T) {
	resolver := &net.Resolver{PreferGo: true}

	for region, targets := range PingTargets {
		for _, target := range targets {
			t.Run(fmt.Sprintf("%s_%s", region, target), func(t *testing.T) {
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				defer cancel()

				addrs, err := resolver.LookupHost(ctx, target)
				if err != nil {
					t.Fatalf("DNS lookup failed for %s: %v", target, err)
				}
				if len(addrs) == 0 {
					t.Fatalf("DNS returned 0 addresses for %s", target)
				}
				t.Logf("OK: %s -> %v", target, addrs[:1])
			})
		}
	}
}

type bytesReader struct {
	data []byte
	pos  int
}

func readerFromBytes(data []byte) io.Reader {
	return &bytesReader{data: data}
}

func (r *bytesReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
