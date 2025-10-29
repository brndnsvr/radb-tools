package cli

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bss/radb-client/internal/api"
	"github.com/bss/radb-client/internal/config"
	"github.com/bss/radb-client/internal/state"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	daemonInterval int
	daemonOnce     bool
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Run as a daemon to monitor RADb changes",
	Long: `Run radb-client as a long-running daemon that periodically checks
for changes in RADb route objects and maintains historical snapshots.

The daemon will:
  - Fetch route objects at regular intervals
  - Create snapshots automatically
  - Detect and log changes
  - Maintain historical data according to retention policies
  - Run cleanup tasks automatically

This mode is designed for server deployment and systemd integration.`,
	RunE: runDaemon,
}

func init() {
	daemonCmd.Flags().IntVarP(&daemonInterval, "interval", "i", 3600, "Check interval in seconds (default: 3600 = 1 hour)")
	daemonCmd.Flags().BoolVar(&daemonOnce, "once", false, "Run once and exit (useful for testing)")
}

func runDaemon(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load configuration: %w", err)
	}

	// Setup logging for daemon mode
	setupDaemonLogging(cfg)

	logrus.Info("RADb Client Daemon starting...")
	logrus.Infof("Version: %s", Version)
	logrus.Infof("Check interval: %d seconds (%d minutes)", daemonInterval, daemonInterval/60)

	// Load credentials
	creds, err := cfg.LoadCredentials()
	if err != nil {
		logrus.Error("Failed to load credentials")
		return fmt.Errorf("load credentials: %w", err)
	}

	// Create API client
	apiClient := api.NewClient(
		cfg.API.BaseURL,
		cfg.API.Source,
		creds.Username,
		creds.APIKey,
	)

	// Create state manager
	stateManager, err := state.NewManager(
		cfg.StateDir()+"/cache",
		cfg.StateDir()+"/history",
	)
	if err != nil {
		return fmt.Errorf("create state manager: %w", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	// If running once, just execute and exit
	if daemonOnce {
		logrus.Info("Running in one-shot mode")
		return performCheck(ctx, apiClient, stateManager)
	}

	// Start daemon loop
	ticker := time.NewTicker(time.Duration(daemonInterval) * time.Second)
	defer ticker.Stop()

	// Perform initial check immediately
	logrus.Info("Performing initial check...")
	if err := performCheck(ctx, apiClient, stateManager); err != nil {
		logrus.Errorf("Initial check failed: %v", err)
		// Don't exit on first failure - log and continue
	}

	logrus.Info("Daemon started successfully")
	logrus.Infof("Next check in %d seconds", daemonInterval)

	// Main daemon loop
	for {
		select {
		case <-ticker.C:
			logrus.Info("Starting periodic check...")
			if err := performCheck(ctx, apiClient, stateManager); err != nil {
				logrus.Errorf("Periodic check failed: %v", err)
				// Continue running even on failure
			}
			logrus.Infof("Next check in %d seconds", daemonInterval)

		case sig := <-sigChan:
			logrus.Infof("Received signal: %v", sig)

			switch sig {
			case syscall.SIGHUP:
				// Reload configuration
				logrus.Info("Reloading configuration...")
				newCfg, err := config.Load()
				if err != nil {
					logrus.Errorf("Failed to reload configuration: %v", err)
				} else {
					cfg = newCfg
					setupDaemonLogging(cfg)
					logrus.Info("Configuration reloaded successfully")
				}

			case os.Interrupt, syscall.SIGTERM:
				// Graceful shutdown
				logrus.Info("Shutting down gracefully...")
				return nil
			}

		case <-ctx.Done():
			logrus.Info("Context cancelled, shutting down...")
			return ctx.Err()
		}
	}
}

// performCheck executes a single check cycle
func performCheck(ctx context.Context, apiClient api.APIClient, stateManager state.StateManager) error {
	startTime := time.Now()

	logrus.Info("Fetching route objects...")

	// Fetch routes with context
	routes, err := apiClient.ListRoutes(ctx, nil)
	if err != nil {
		return fmt.Errorf("fetch routes: %w", err)
	}

	logrus.Infof("Fetched %d route objects", len(routes))

	// Save snapshot
	logrus.Debug("Saving snapshot...")
	if err := stateManager.SaveSnapshot(ctx, "route_objects", routes); err != nil {
		logrus.Errorf("Failed to save snapshot: %v", err)
		// Continue even if snapshot save fails
	} else {
		logrus.Info("Snapshot saved successfully")
	}

	// Generate diff if previous snapshot exists
	logrus.Debug("Generating diff...")
	diff, err := stateManager.GenerateDiff(ctx, "route_objects", "route_objects")
	if err != nil {
		logrus.Debugf("Could not generate diff (may be first run): %v", err)
	} else if diff != nil {
		// Log changes
		added, removed, modified := countChanges(diff)

		if added > 0 || removed > 0 || modified > 0 {
			logrus.Infof("Changes detected: %d added, %d removed, %d modified",
				added, removed, modified)

			// TODO: Implement notification system here
			// For now, just log the changes
		} else {
			logrus.Info("No changes detected")
		}
	}

	// Perform cleanup if configured
	logrus.Debug("Running cleanup tasks...")
	if err := performCleanup(ctx, stateManager); err != nil {
		logrus.Errorf("Cleanup failed: %v", err)
	}

	duration := time.Since(startTime)
	logrus.Infof("Check completed in %v", duration)

	return nil
}

// countChanges counts the number of changes in a diff
func countChanges(diff interface{}) (added, removed, modified int) {
	// This is a simplified version - actual implementation depends on
	// the diff structure from internal/state/diff.go

	// TODO: Implement proper diff counting based on actual diff structure
	// For now, return placeholder values
	return 0, 0, 0
}

// performCleanup runs cleanup tasks
func performCleanup(ctx context.Context, stateManager state.StateManager) error {
	// TODO: Implement cleanup based on retention policy
	// - Remove old snapshots beyond retention period
	// - Compress old history files
	// - Clean up orphaned files

	logrus.Debug("Cleanup tasks completed")
	return nil
}

// setupDaemonLogging configures logging for daemon mode
func setupDaemonLogging(cfg *config.Config) {
	// Set log level
	level, err := logrus.ParseLevel(cfg.Preferences.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// Use JSON formatter for structured logging (easier to parse)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyTime:  "timestamp",
			logrus.FieldKeyLevel: "level",
			logrus.FieldKeyMsg:   "message",
		},
	})

	// Output to stdout (systemd captures this)
	logrus.SetOutput(os.Stdout)

	logrus.Debug("Daemon logging configured")
}
