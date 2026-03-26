package commands

import "flag"

var UploadCmd = flag.NewFlagSet("upload", flag.ExitOnError)

func init() {
	UploadCmd.String("url", "", "Upload endpoint URL (uses default servers if empty)")
	UploadCmd.Int("concurrency", 4, "Number of concurrent uploads (default: 4)")
	UploadCmd.Int("duration", 10, "Test duration in seconds")
	UploadCmd.String("region", "auto", "Server region: cn, global, auto")
	UploadCmd.Bool("verbose", false, "Enable detailed output")
}
