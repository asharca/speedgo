package commands

import (
	"flag"
)

var PingCmd = flag.NewFlagSet("ping", flag.ExitOnError)

func init() {
	PingCmd.String("targets", "", "Comma-separated list of targets (uses defaults per region if empty)")
	PingCmd.Int("count", 4, "Number of pings per target (default: 4)")
	PingCmd.Duration("timeout", 1_000_000_000, "Timeout for each ping (e.g., 1s, 500ms)")
	PingCmd.Int("concurrency", 3, "Number of concurrent pings (default: 3)")
	PingCmd.String("region", "auto", "Server region: cn, global, auto")
	PingCmd.Bool("verbose", false, "Enable detailed output")
}
