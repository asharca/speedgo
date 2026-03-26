package core

import (
	"testing"
	"time"
)

func TestParseAllConfig(t *testing.T) {
	config, err := parseAllConfig([]string{"-region=cn", "-concurrency=2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionCN {
		t.Errorf("Region = %s, want cn", config.Region)
	}
	if config.Concurrency != 2 {
		t.Errorf("Concurrency = %d, want 2", config.Concurrency)
	}
}

func TestParseAllConfig_Global(t *testing.T) {
	config, err := parseAllConfig([]string{"-region=global", "-concurrency=8"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if config.Region != RegionGlobal {
		t.Errorf("Region = %s, want global", config.Region)
	}
	if config.Concurrency != 8 {
		t.Errorf("Concurrency = %d, want 8", config.Concurrency)
	}
}

func TestPrintSummary(t *testing.T) {
	results := AllResults{
		PingResults: []PingResult{
			{
				Target: "test.com",
				RTTs:   []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				MinRTT: 10 * time.Millisecond,
				MaxRTT: 20 * time.Millisecond,
				AvgRTT: 15 * time.Millisecond,
			},
		},
		DownloadStat: DownloadStats{
			BytesReceived: 10000000,
			Duration:      5 * time.Second,
			Speed:         16.0,
		},
		UploadStat: UploadStats{
			BytesSent: 5000000,
			Duration:  5 * time.Second,
			Speed:     8.0,
		},
	}

	// Should not panic
	printSummary(results)
}

func TestPrintSummary_EmptyPing(t *testing.T) {
	results := AllResults{
		DownloadStat: DownloadStats{Speed: 10.0},
		UploadStat:   UploadStats{Speed: 5.0},
	}
	// Should not panic with empty ping results
	printSummary(results)
}

func TestPrintSummary_AllZeroPing(t *testing.T) {
	results := AllResults{
		PingResults: []PingResult{
			{Target: "fail.com", Lost: 4},
		},
		DownloadStat: DownloadStats{Speed: 10.0},
		UploadStat:   UploadStats{Speed: 5.0},
	}
	// Should not panic when all pings failed
	printSummary(results)
}
