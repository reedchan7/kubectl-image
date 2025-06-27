package main

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// These variables will be set at build time via ldflags
var (
	version   = "dirty"
	commit    = "unknown"
	date      = "unknown"
	goVersion = runtime.Version()
)

// createVersionCommand creates the 'version' subcommand
func createVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long: `Print version information for kubectl-image plugin.

Examples:
  # Show version information
  kubectl image version
`,
		Run: func(cmd *cobra.Command, args []string) {
			runVersionCommand()
		},
	}

	return cmd
}

// runVersionCommand handles the version command execution
func runVersionCommand() {
	fmt.Printf("kubectl-image version %s\n", version)
	fmt.Printf("  commit: %s\n", commit)
	fmt.Printf("  built: %s\n", date)
	fmt.Printf("  go version: %s\n", goVersion)
	fmt.Printf("  platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
}
