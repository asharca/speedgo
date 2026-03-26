// core/upload.go
package core

import (
	"bytes"
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"net/http"
	"speedgo/commands"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type UploadConfig struct {
	Duration    time.Duration
	Concurrency int
	Region      Region
	Verbose     bool
}

type UploadStats struct {
	BytesSent int64
	Duration  time.Duration
	Speed     float64
	Error     error
}

const (
	chunkSize = 1 * 1024 * 1024 // 1MB chunks
)

// getUploadEndpoint returns the upload URL based on region.
func getUploadEndpoint(region Region) string {
	servers := GetUploadServers(region)
	if len(servers) > 0 {
		return servers[0].URL
	}
	return "https://speed.cloudflare.com/__up"
}

func RunUpload(ctx context.Context, args []string) error {
	config, err := parseUploadConfig(args)
	if err != nil {
		return fmt.Errorf("parsing upload config: %w", err)
	}

	// Auto-select best upload server
	candidates := GetUploadServers(config.Region)
	best := SelectBestServers(ctx, candidates, 1)
	endpoint := best[0].URL

	fmt.Printf("\nStarting upload speed test (Duration: %v, Streams: %d, Server: %s)\n",
		config.Duration, config.Concurrency, best[0].Name)

	stats := measureUploadSpeed(ctx, config, endpoint)
	printUploadResults(stats)

	return nil
}

func measureUploadSpeed(ctx context.Context, config *UploadConfig, endpoint string) UploadStats {
	var totalBytes int64
	start := time.Now()

	// Create channels for coordination
	errChan := make(chan error, config.Concurrency)
	bytesChan := make(chan int64, config.Concurrency)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, config.Duration)
	defer cancel()

	// Generate test data
	testData := generateTestData(chunkSize)

	// Start concurrent uploads
	var wg sync.WaitGroup
	for i := 0; i < config.Concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			uploadWorker(ctx, config, endpoint, testData, bytesChan, errChan)
		}(i)
	}

	// Start progress display
	progress := NewProgress(&totalBytes, "Upload  ")
	defer progress.Stop()

	// Collect results
	go func() {
		wg.Wait()
		close(bytesChan)
		close(errChan)
	}()

	// Process results
	var lastError error
	for {
		select {
		case bytes, ok := <-bytesChan:
			if !ok {
				duration := time.Since(start)
				return UploadStats{
					BytesSent: totalBytes,
					Duration:  duration,
					Speed:     float64(totalBytes*8) / (1000 * 1000 * duration.Seconds()),
					Error:     lastError,
				}
			}
			atomic.AddInt64(&totalBytes, bytes)

		case err := <-errChan:
			if err != nil {
				lastError = err
			}
		}
	}
}

func uploadWorker(ctx context.Context, _ *UploadConfig, endpoint string,
	testData []byte, bytesChan chan<- int64, errChan chan<- error) {

	client := &http.Client{
		Timeout: 10 * time.Second, // Individual request timeout
		Transport: &http.Transport{
			MaxIdleConns:       100,
			IdleConnTimeout:    90 * time.Second,
			DisableCompression: true,
			MaxConnsPerHost:    100,
		},
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if err := uploadChunk(ctx, client, endpoint, testData, bytesChan); err != nil {
				errChan <- fmt.Errorf("upload error: %w", err)
				time.Sleep(100 * time.Millisecond) // Short backoff on error
				continue
			}
		}
	}
}

func uploadChunk(ctx context.Context, client *http.Client, endpoint string, data []byte, bytesChan chan<- int64) error {
	reader := &countingReader{
		reader: bytes.NewReader(data),
		count:  0,
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, reader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	SetUA(req)
	req.Header.Set("Content-Type", "application/octet-stream")
	req.Header.Set("Content-Length", fmt.Sprint(len(data)))

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(body))
	}

	// Report bytes uploaded
	bytesChan <- reader.count
	return nil
}

type countingReader struct {
	reader io.Reader
	count  int64
}

func (r *countingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	atomic.AddInt64(&r.count, int64(n))
	return n, err
}

func generateTestData(size int) []byte {
	data := make([]byte, size)
	if _, err := rand.Read(data); err != nil {
		// Fall back to predictable pattern if random fails
		for i := range data {
			data[i] = byte(i % 256)
		}
	}
	return data
}

func parseUploadConfig(args []string) (*UploadConfig, error) {
	cmd := commands.UploadCmd
	if err := cmd.Parse(args); err != nil {
		return nil, fmt.Errorf("parsing arguments: %w", err)
	}

	duration := cmd.Lookup("duration").Value.(flag.Getter).Get().(int)

	return &UploadConfig{
		Duration:    time.Duration(duration) * time.Second,
		Concurrency: cmd.Lookup("concurrency").Value.(flag.Getter).Get().(int),
		Region:      Region(cmd.Lookup("region").Value.String()),
		Verbose:     cmd.Lookup("verbose").Value.(flag.Getter).Get().(bool),
	}, nil
}

func printUploadResults(stats UploadStats) {
	fmt.Printf("\n\nUPLOAD TEST RESULTS\n")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Total data sent: %.2f MB\n", float64(stats.BytesSent)/(1000*1000))
	fmt.Printf("Test duration: %.1f seconds\n", stats.Duration.Seconds())
	fmt.Printf("Average speed: %.2f Mbps\n", stats.Speed)
	if stats.Error != nil {
		fmt.Printf("Errors encountered: %v\n", stats.Error)
	}
	fmt.Println(strings.Repeat("=", 50))
}
