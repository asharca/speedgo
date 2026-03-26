package core

import (
	"errors"
	"testing"
	"time"
)

func TestPrintDownloadResults(t *testing.T) {
	stats := DownloadStats{
		BytesReceived: 50000000,
		Duration:      10 * time.Second,
		Speed:         40.0,
	}
	printDownloadResults(stats) // Should not panic
}

func TestPrintDownloadResults_WithError(t *testing.T) {
	stats := DownloadStats{
		BytesReceived: 1000,
		Duration:      1 * time.Second,
		Speed:         0.008,
		Error:         errors.New("test error"),
	}
	printDownloadResults(stats) // Should not panic
}

func TestPrintUploadResults(t *testing.T) {
	stats := UploadStats{
		BytesSent: 25000000,
		Duration:  10 * time.Second,
		Speed:     20.0,
	}
	printUploadResults(stats) // Should not panic
}

func TestPrintUploadResults_WithError(t *testing.T) {
	stats := UploadStats{
		BytesSent: 500,
		Duration:  1 * time.Second,
		Speed:     0.004,
		Error:     errors.New("upload error"),
	}
	printUploadResults(stats) // Should not panic
}

func TestPrintResults_Ping(t *testing.T) {
	results := []PingResult{
		{
			Target: "google.com",
			RTTs:   []time.Duration{10 * time.Millisecond, 20 * time.Millisecond, 15 * time.Millisecond},
			MinRTT: 10 * time.Millisecond,
			MaxRTT: 20 * time.Millisecond,
			AvgRTT: 15 * time.Millisecond,
			Lost:   0,
		},
		{
			Target: "failed.com",
			RTTs:   nil,
			Lost:   4,
			Errors: []error{errors.New("timeout")},
		},
	}
	printResults(results) // Should not panic
}

func TestPrintResults_Empty(t *testing.T) {
	printResults(nil) // Should not panic
}
