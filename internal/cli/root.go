// Package cli provides the command-line interface for the RADb client.
package cli

import (
	"fmt"
	"os"

	"github.com/bss/radb-client/internal/api"
	"github.com/bss/radb-client/internal/config"
	"github.com/bss/radb-client/internal/state"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Global context shared across commands
type CLIContext struct {
	Config     *config.Config
	APIClient  api.Client
	StateMgr   state.Manager
	CredMgr    *config.CredentialManager
	Logger     *logrus.Logger
}

var (
	ctx CLIContext

	rootCmd = &cobra.Command{
		Use:   "radb-client",
		Short: "RADb API client for route and contact management",
		Long: `A command-line client for interacting with the RADb (Routing Assets Database) API.
Manage route objects, contacts, and track changes over time.`,
		PersistentPreRunE: initializeContext,
		SilenceUsage:      true,
		SilenceErrors:     true,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("config", "", "config file (default is $HOME/.radb-client/config.yaml)")
	rootCmd.PersistentFlags().Bool("debug", false, "enable debug logging")

	// Create logger for command initialization
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Add subcommands
	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(NewWizardCmd(logger))

	// Phase 2 commands
	rootCmd.AddCommand(NewRouteCmd(logger))
	rootCmd.AddCommand(NewContactCmd(logger))
	rootCmd.AddCommand(NewSnapshotCmd(logger))

	// Phase 3 commands
	rootCmd.AddCommand(NewHistoryCmd(logger))
	rootCmd.AddCommand(NewSearchCmd(logger))
}

// initializeContext initializes the CLI context before command execution.
func initializeContext(cmd *cobra.Command, args []string) error {
	// Skip initialization for certain commands
	skipInit := []string{"config init", "version", "help"}
	for _, skip := range skipInit {
		if cmd.CommandPath() == "radb-client "+skip {
			return nil
		}
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w (try running 'radb-client config init')", err)
	}

	// Setup logger
	logger := cfg.GetLogger()
	if debug, _ := cmd.Flags().GetBool("debug"); debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	ctx.Config = cfg
	ctx.Logger = logger

	// Initialize credential manager
	credMgr, err := config.NewCredentialManager(cfg.ConfigDir, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize credential manager: %w", err)
	}
	ctx.CredMgr = credMgr

	// Initialize API client
	ctx.APIClient = api.NewHTTPClient(
		cfg.API.BaseURL,
		cfg.API.Source,
		cfg.API.Timeout,
		logger,
	)

	// Initialize state manager
	stateMgr, err := state.NewFileManager(cfg.Preferences.CacheDir, logger)
	if err != nil {
		return fmt.Errorf("failed to initialize state manager: %w", err)
	}
	ctx.StateMgr = stateMgr

	return nil
}

// cleanup performs cleanup operations on exit.
func cleanup() {
	if ctx.StateMgr != nil {
		ctx.StateMgr.Close()
	}
	if ctx.CredMgr != nil {
		ctx.CredMgr.Close()
	}
}

// handleError handles command errors with appropriate logging and exit codes.
func handleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		cleanup()
		os.Exit(1)
	}
}
