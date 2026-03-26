package core

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
)

// ProbeResult holds latency probe result for a server.
type ProbeResult struct {
	Server  TestServer
	Latency time.Duration
	Error   error
}

// ProbeServers tests latency to all given servers and returns results sorted by latency.
// Servers that fail the probe are placed at the end.
func ProbeServers(ctx context.Context, servers []TestServer, timeout time.Duration) []ProbeResult {
	results := make([]ProbeResult, len(servers))
	var wg sync.WaitGroup

	for i, server := range servers {
		wg.Add(1)
		go func(idx int, s TestServer) {
			defer wg.Done()
			latency, err := probeLatency(ctx, s.URL, timeout)
			results[idx] = ProbeResult{
				Server:  s,
				Latency: latency,
				Error:   err,
			}
		}(i, server)
	}

	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		if results[i].Error != nil && results[j].Error != nil {
			return false
		}
		if results[i].Error != nil {
			return false
		}
		if results[j].Error != nil {
			return true
		}
		return results[i].Latency < results[j].Latency
	})

	return results
}

// probeLatency sends a small HTTP request to measure round-trip time.
func probeLatency(ctx context.Context, rawURL string, timeout time.Duration) (time.Duration, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Extract just the origin for the probe (don't download the full file)
	u, err := url.Parse(rawURL)
	if err != nil {
		return 0, err
	}
	probeURL := fmt.Sprintf("%s://%s/", u.Scheme, u.Host)

	req, err := http.NewRequestWithContext(ctx, "HEAD", probeURL, nil)
	if err != nil {
		return 0, err
	}

	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Don't follow redirects
		},
	}

	start := time.Now()
	resp, err := client.Do(req)
	latency := time.Since(start)

	if err != nil {
		return 0, err
	}
	resp.Body.Close()

	return latency, nil
}

// SelectBestServers probes servers and returns up to n best ones.
func SelectBestServers(ctx context.Context, servers []TestServer, n int) []TestServer {
	if len(servers) == 0 {
		return nil
	}

	fmt.Printf("Probing %d servers for latency...\n", len(servers))
	results := ProbeServers(ctx, servers, 5*time.Second)

	var best []TestServer
	for i, r := range results {
		if i >= n {
			break
		}
		if r.Error != nil {
			continue
		}
		fmt.Printf("  %-20s %v\n", r.Server.Name, r.Latency.Round(time.Millisecond))
		best = append(best, r.Server)
	}

	if len(best) == 0 {
		fmt.Println("  No servers reachable, using all as fallback")
		return servers
	}

	return best
}
