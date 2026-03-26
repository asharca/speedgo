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

	fmt.Println()
	fmt.Printf("  %s  %s\n",
		colorize(bold+cyan, "SpeedGo"),
		colorf(dim, "region=%s", config.Region))
	fmt.Printf("  %s\n", colorize(dim, strings.Repeat("─", 50)))

	var results AllResults

	// 1. Ping test
	fmt.Printf("\n  %s\n", colorf(bold, "[1/3] Latency"))
	pingArgs := []string{
		fmt.Sprintf("-region=%s", config.Region),
		"-count=3",
		fmt.Sprintf("-concurrency=%d", config.Concurrency),
	}
	pingConfig, err := NewPingConfig(pingArgs)
	if err != nil {
		fmt.Printf("  %s %v\n", colorize(yellow, "!"), err)
	} else {
		results.PingResults = pingTargets(ctx, pingConfig)
		printResults(results.PingResults)
	}

	// 2. Download test
	fmt.Printf("\n  %s\n", colorf(bold, "[2/3] Download"))
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
	fmt.Printf("\n  %s\n", colorf(bold, "[3/3] Upload"))
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

func speedRating(mbps float64) string {
	switch {
	case mbps >= 100:
		return colorf(green, "%.2f Mbps", mbps)
	case mbps >= 30:
		return colorf(yellow, "%.2f Mbps", mbps)
	default:
		return colorf(red, "%.2f Mbps", mbps)
	}
}

func printSummary(results AllResults) {
	fmt.Println()
	fmt.Printf("  %s\n", colorize(dim, strings.Repeat("─", 50)))
	fmt.Printf("  %s\n", colorize(bold+cyan, "RESULTS"))
	fmt.Println()

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
			ms := float64(bestPing.Microseconds()) / 1000
			fmt.Printf("  %s  %s  %s\n",
				colorize(dim, "Latency "),
				colorf(rttColor(ms)+bold, "%.1f ms", ms),
				colorf(dim, "(%s)", bestTarget))
		}
	}

	fmt.Printf("  %s  %s\n",
		colorize(dim, "Download"),
		colorize(bold, speedRating(results.DownloadStat.Speed)))
	fmt.Printf("  %s  %s\n",
		colorize(dim, "Upload  "),
		colorize(bold, speedRating(results.UploadStat.Speed)))

	fmt.Printf("\n  %s\n\n", colorize(dim, strings.Repeat("─", 50)))
}
