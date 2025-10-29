// Package main is the entry point for the RADb client application.
package main

import (
	"os"

	"github.com/bss/radb-client/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		// Error is already printed by cobra (if SilenceErrors is false)
		// or by the command itself
		os.Exit(1)
	}
}
