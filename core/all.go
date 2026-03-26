package core

import (
	"context"
	"flag"
	"fmt"
	"speedgo/commands"
	"strings"
	"time"
)

// AllConfig stores configuration for the full speed test.
type AllConfig struct {
	Region      Region
	Concurrency int
	Verbose     bool
}

// AllResults stores combined results from all tests.
type AllResults struct {
	PingResults  []PingResult
	DownloadStat DownloadStats
	UploadStat   UploadStats
}

func RunAll(ctx context.Context, args []string) error {
	config, err := parseAllConfig(args)
	if err != nil {
		return fmt.Errorf("parsing config: %w", err)
	}

	fmt.Printf("Running full speed test (Region: %s)\n", config.Region)
	fmt.Println(strings.Repeat("=", 50))

	var results AllResults

	// 1. Ping test
	fmt.Println("\n[1/3] Latency Test")
	pingArgs := []string{
		fmt.Sprintf("-region=%s", config.Region),
		"-count=3",
		fmt.Sprintf("-concurrency=%d", config.Concurrency),
	}
	pingConfig, err := NewPingConfig(pingArgs)
	if err != nil {
		fmt.Printf("  Ping skipped: %v\n", err)
	} else {
		results.PingResults = pingTargets(ctx, pingConfig)
		printResults(results.PingResults)
	}

	// 2. Download test
	fmt.Println("\n[2/3] Download Test")
	dlArgs := []string{
		fmt.Sprintf("-region=%s", config.Region),
		"-duration=15s",
		fmt.Sprintf("-concurrency=%d", config.Concurrency),
	}
	dlConfig, err := parseDownloadConfig(dlArgs)
	if err != nil {
		return fmt.Errorf("download config: %w", err)
	}
	candidates := GetDownloadServers(dlConfig.Region)
	best := SelectBestServers(ctx, candidates, dlConfig.Concurrency)
	dlConfig.servers = best
	results.DownloadStat = measureDownloadSpeed(ctx, dlConfig)
	printDownloadResults(results.DownloadStat)

	// 3. Upload test
	fmt.Println("\n[3/3] Upload Test")
	ulArgs := []string{
		fmt.Sprintf("-region=%s", config.Region),
		"-duration=10",
		fmt.Sprintf("-concurrency=%d", config.Concurrency),
	}
	ulConfig, err := parseUploadConfig(ulArgs)
	if err != nil {
		return fmt.Errorf("upload config: %w", err)
	}
	ulCandidates := GetUploadServers(ulConfig.Region)
	ulBest := SelectBestServers(ctx, ulCandidates, 1)
	endpoint := ulBest[0].URL
	results.UploadStat = measureUploadSpeed(ctx, ulConfig, endpoint)
	printUploadResults(results.UploadStat)

	// Summary
	printSummary(results)

	return nil
}

func parseAllConfig(args []string) (*AllConfig, error) {
	cmd := commands.AllCmd
	if err := cmd.Parse(args); err != nil {
		return nil, err
	}
	return &AllConfig{
		Region:      Region(cmd.Lookup("region").Value.String()),
		Concurrency: cmd.Lookup("concurrency").Value.(flag.Getter).Get().(int),
		Verbose:     cmd.Lookup("verbose").Value.(flag.Getter).Get().(bool),
	}, nil
}

func printSummary(results AllResults) {
	fmt.Printf("\n\nSPEED TEST SUMMARY\n")
	fmt.Println(strings.Repeat("=", 50))

	// Best ping
	if len(results.PingResults) > 0 {
		var bestPing time.Duration
		var bestTarget string
		for _, r := range results.PingResults {
			if r.AvgRTT > 0 && (bestPing == 0 || r.AvgRTT < bestPing) {
				bestPing = r.AvgRTT
				bestTarget = r.Target
			}
		}
		if bestPing > 0 {
			fmt.Printf("  Latency:  %.1f ms (%s)\n",
				float64(bestPing.Microseconds())/1000, bestTarget)
		}
	}

	fmt.Printf("  Download: %.2f Mbps\n", results.DownloadStat.Speed)
	fmt.Printf("  Upload:   %.2f Mbps\n", results.UploadStat.Speed)
	fmt.Println(strings.Repeat("=", 50))
}
