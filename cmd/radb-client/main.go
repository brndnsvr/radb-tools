// Package main is the entry point for the RADb client application.
package main

import (
	"os"

	"github.com/bss/radb-client/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		os.Exit(1)
	}
}
