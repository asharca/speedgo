package main

import (
	"context"
	"fmt"
	"os"
	"speedgo/commands"
	"speedgo/core"
)

func main() {
	ctx := context.Background()

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	switch cmd {
	case "ping", "p":
		if err := pingCommand(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", core.Colorize(core.Red, "Error:"), err)
			os.Exit(1)
		}
	case "download", "d":
		if err := downloadCommand(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", core.Colorize(core.Red, "Error:"), err)
			os.Exit(1)
		}
	case "upload", "u":
		if err := uploadCommand(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", core.Colorize(core.Red, "Error:"), err)
			os.Exit(1)
		}
	case "all", "a":
		if err := allCommand(ctx, args); err != nil {
			fmt.Fprintf(os.Stderr, "%s %v\n", core.Colorize(core.Red, "Error:"), err)
			os.Exit(1)
		}
	case "-h", "--help", "help":
		printHelp()
	case "-v", "--version", "version":
		fmt.Printf("%s %s\n", core.Colorize(core.Bold+core.Cyan, "SpeedGo"), core.Colorize(core.Dim, "v1.0.0"))
	default:
		fmt.Fprintf(os.Stderr, "%s unknown command: %s\n\n", core.Colorize(core.Red, "Error:"), cmd)
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	c := core.Colorize
	f := core.Colorf

	fmt.Println()
	fmt.Printf("  %s %s\n", c(core.Bold+core.Cyan, "SpeedGo"), c(core.Dim, "— Network speed test CLI"))
	fmt.Println()

	fmt.Printf("  %s\n", c(core.Bold, "USAGE"))
	fmt.Printf("    %s %s %s\n",
		c(core.White, "speedgo"),
		c(core.Cyan, "<command>"),
		c(core.Dim, "[options]"))
	fmt.Println()

	fmt.Printf("  %s\n", c(core.Bold, "COMMANDS"))
	fmt.Printf("    %s  %s\n", f(core.Cyan, "%-14s", "all, a"), c(core.Dim, "Run full speed test (ping + download + upload)"))
	fmt.Printf("    %s  %s\n", f(core.Cyan, "%-14s", "ping, p"), c(core.Dim, "Test network latency"))
	fmt.Printf("    %s  %s\n", f(core.Cyan, "%-14s", "download, d"), c(core.Dim, "Test download speed"))
	fmt.Printf("    %s  %s\n", f(core.Cyan, "%-14s", "upload, u"), c(core.Dim, "Test upload speed"))
	fmt.Printf("    %s  %s\n", f(core.Cyan, "%-14s", "version"), c(core.Dim, "Show version"))
	fmt.Println()

	fmt.Printf("  %s\n", c(core.Bold, "OPTIONS"))
	fmt.Printf("    %s  %s  %s\n",
		f(core.Green, "%-14s", "--region"), c(core.Dim, "cn | global | auto"), c(core.Dim, "(default: auto)"))
	fmt.Printf("    %s  %s\n",
		f(core.Green, "%-14s", "--concurrency"), c(core.Dim, "Number of parallel streams (default: 2)"))
	fmt.Printf("    %s  %s\n",
		f(core.Green, "%-14s", "--verbose"), c(core.Dim, "Show detailed output"))
	fmt.Println()

	fmt.Printf("  %s\n", c(core.Bold, "EXAMPLES"))
	fmt.Printf("    %s    %s\n", c(core.White, "speedgo all"), c(core.Dim, "# Full test, auto region"))
	fmt.Printf("    %s    %s\n", c(core.White, "speedgo all --region=cn"), c(core.Dim, "# Full test, CN servers"))
	fmt.Printf("    %s    %s\n", c(core.White, "speedgo d --region=global"), c(core.Dim, "# Download, global servers"))
	fmt.Printf("    %s    %s\n", c(core.White, "speedgo ping --targets=google.com"), c(core.Dim, "# Ping custom target"))
	fmt.Println()
}

func pingCommand(ctx context.Context, args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		commands.PingCmd.Usage()
		return nil
	}
	return core.RunPing(ctx, args)
}

func downloadCommand(ctx context.Context, args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		commands.DownloadCmd.Usage()
		return nil
	}
	return core.RunDownload(ctx, args)
}

func uploadCommand(ctx context.Context, args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		commands.UploadCmd.Usage()
		return nil
	}
	return core.RunUpload(ctx, args)
}

func allCommand(ctx context.Context, args []string) error {
	if len(args) > 0 && (args[0] == "-h" || args[0] == "--help") {
		commands.AllCmd.Usage()
		return nil
	}
	return core.RunAll(ctx, args)
}
