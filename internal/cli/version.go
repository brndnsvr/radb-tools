package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	// Version is the current version of the application
	Version = "0.9.0-pre"

	// BuildDate is set during build
	BuildDate = "development"

	// GitCommit is set during build
	GitCommit = "dev"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, build date, and git commit information.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("radb-client version %s\n", Version)
		fmt.Printf("Build date: %s\n", BuildDate)
		fmt.Printf("Git commit: %s\n", GitCommit)
		fmt.Println("\n🧪 Pre-release build - pending final manual testing")
		fmt.Println("\nSee TESTING_RUNBOOK.md for complete testing procedures")
	},
}
