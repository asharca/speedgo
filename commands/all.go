package commands

import "flag"

var AllCmd = flag.NewFlagSet("all", flag.ExitOnError)

func init() {
	AllCmd.String("region", "auto", "Server region: cn, global, auto")
	AllCmd.Int("concurrency", 4, "Number of concurrent streams")
	AllCmd.Bool("verbose", false, "Enable detailed output")
}
