package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/bss/radb-client/internal/config"
	"github.com/bss/radb-client/internal/state"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewHistoryCmd creates the history command and its subcommands.
func NewHistoryCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "history",
		Aliases: []string{"hist"},
		Short:   "View change history",
		Long:    "View and analyze change history from the changelog",
	}

	cmd.AddCommand(
		newHistoryShowCmd(logger),
		newHistoryStatsCmd(logger),
	)

	return cmd
}

// newHistoryShowCmd creates the history show command.
func newHistoryShowCmd(logger *logrus.Logger) *cobra.Command {
	var (
		outputFormat string
		since        string
		until        string
		objectType   string
		limit        int
	)

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show change history",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			historyMgr := state.NewHistoryManager(cfg.StateDir(), logger)

			// Parse time range
			var fromTime, toTime time.Time
			if since != "" {
				fromTime, err = parseTimeSpec(since)
				if err != nil {
					return fmt.Errorf("invalid since time: %w", err)
				}
			} else {
				fromTime = time.Now().Add(-30 * 24 * time.Hour) // Default: last 30 days
			}

			if until != "" {
				toTime, err = parseTimeSpec(until)
				if err != nil {
					return fmt.Errorf("invalid until time: %w", err)
				}
			} else {
				toTime = time.Now()
			}

			// Query changes
			entries, err := historyMgr.QueryChanges(ctx, fromTime, toTime, objectType)
			if err != nil {
				return fmt.Errorf("failed to query history: %w", err)
			}

			// Apply limit if specified
			if limit > 0 && len(entries) > limit {
				entries = entries[len(entries)-limit:]
			}

			// Render output
			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			return outputter.RenderChangeHistory(entries)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVar(&since, "since", "", "Show changes since (e.g., '2024-01-01', '7d', '1h')")
	cmd.Flags().StringVar(&until, "until", "", "Show changes until (e.g., '2024-12-31')")
	cmd.Flags().StringVar(&objectType, "type", "", "Filter by object type (route, contact)")
	cmd.Flags().IntVar(&limit, "limit", 0, "Limit number of entries shown")

	return cmd
}

// newHistoryStatsCmd creates the history stats command.
func newHistoryStatsCmd(logger *logrus.Logger) *cobra.Command {
	var (
		outputFormat string
		since        string
		until        string
	)

	cmd := &cobra.Command{
		Use:   "stats",
		Short: "Show change statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			historyMgr := state.NewHistoryManager(cfg.StateDir(), logger)

			// Parse time range
			var fromTime, toTime time.Time
			if since != "" {
				fromTime, err = parseTimeSpec(since)
				if err != nil {
					return fmt.Errorf("invalid since time: %w", err)
				}
			} else {
				fromTime = time.Now().Add(-30 * 24 * time.Hour)
			}

			if until != "" {
				toTime, err = parseTimeSpec(until)
				if err != nil {
					return fmt.Errorf("invalid until time: %w", err)
				}
			} else {
				toTime = time.Now()
			}

			// Get statistics
			stats, err := historyMgr.GetStatistics(ctx, fromTime, toTime)
			if err != nil {
				return fmt.Errorf("failed to get statistics: %w", err)
			}

			// Render output
			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			switch outputFormat {
			case "json":
				return outputter.renderJSON(stats)
			case "yaml":
				return outputter.renderYAML(stats)
			default:
				fmt.Printf("Change Statistics\n")
				fmt.Printf("=================\n\n")
				fmt.Printf("Time Range: %s to %s\n\n",
					stats.TimeRange.From.Format("2006-01-02"),
					stats.TimeRange.To.Format("2006-01-02"))
				fmt.Printf("Total Changes: %d\n\n", stats.TotalChanges)

				fmt.Printf("By Type:\n")
				for changeType, count := range stats.ByType {
					fmt.Printf("  %s: %d\n", changeType, count)
				}

				fmt.Printf("\nBy Object Type:\n")
				for objType, count := range stats.ByObjectType {
					fmt.Printf("  %s: %d\n", objType, count)
				}

				if !stats.FirstChange.IsZero() {
					fmt.Printf("\nFirst Change: %s\n", stats.FirstChange.Format("2006-01-02 15:04:05"))
					fmt.Printf("Last Change: %s\n", stats.LastChange.Format("2006-01-02 15:04:05"))
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().StringVar(&since, "since", "", "Statistics since (e.g., '2024-01-01', '7d')")
	cmd.Flags().StringVar(&until, "until", "", "Statistics until (e.g., '2024-12-31')")

	return cmd
}

// parseTimeSpec parses various time specifications.
func parseTimeSpec(spec string) (time.Time, error) {
	// Try parsing as duration relative to now
	if d, err := time.ParseDuration(spec); err == nil {
		return time.Now().Add(-d), nil
	}

	// Try parsing as absolute date/time
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		time.RFC3339,
	}

	for _, format := range formats {
		if t, err := time.Parse(format, spec); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse time specification: %s", spec)
}
