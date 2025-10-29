package cli

import (
	"encoding/json"
	"fmt"

	"github.com/bss/radb-client/internal/version"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	versionShort  bool
	versionFormat string
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  "Display version, build date, and git commit information.",
	Run: func(cmd *cobra.Command, args []string) {
		// Short version flag
		if versionShort {
			fmt.Println(version.Short())
			return
		}

		// Handle different output formats
		switch versionFormat {
		case "json":
			data, _ := json.MarshalIndent(version.Get(), "", "  ")
			fmt.Println(string(data))

		case "yaml":
			data, _ := yaml.Marshal(version.Get())
			fmt.Print(string(data))

		default: // text format
			fmt.Println(version.Full())

			// Show pre-release warning if applicable
			if version.IsPreRelease() {
				fmt.Println("\n🧪 Pre-release build - pending final manual testing")
				fmt.Println("\nSee TESTING_RUNBOOK.md for complete testing procedures")
			}
		}
	},
}

func init() {
	versionCmd.Flags().BoolVarP(&versionShort, "short", "s", false, "Show only version number")
	versionCmd.Flags().StringVarP(&versionFormat, "output", "o", "text", "Output format (text, json, yaml)")
}
