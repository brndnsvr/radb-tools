package cli

import (
	"context"
	"fmt"

	"github.com/bss/radb-client/internal/config"
	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/internal/state"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewSnapshotCmd creates the snapshot command and its subcommands.
func NewSnapshotCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "snapshot",
		Aliases: []string{"snap", "snapshots"},
		Short:   "Manage snapshots",
		Long:    "Create, list, and manage snapshots of routes and contacts",
	}

	cmd.AddCommand(
		newSnapshotCreateCmd(logger),
		newSnapshotListCmd(logger),
		newSnapshotShowCmd(logger),
		newSnapshotDeleteCmd(logger),
	)

	return cmd
}

// newSnapshotCreateCmd creates the snapshot create command.
func newSnapshotCreateCmd(logger *logrus.Logger) *cobra.Command {
	var (
		snapshotType string
		note         string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
			defer stateManager.Close()

			// For now, create an empty snapshot
			// In a real implementation, this would fetch current data from the API
			snapshot := models.NewSnapshot(models.SnapshotType(snapshotType), note)

			if err := snapshot.ComputeChecksum(); err != nil {
				return fmt.Errorf("failed to compute checksum: %w", err)
			}

			if err := stateManager.SaveSnapshot(ctx, snapshot); err != nil {
				return fmt.Errorf("failed to save snapshot: %w", err)
			}

			fmt.Printf("Created snapshot: %s\n", snapshot.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&snapshotType, "type", "route", "Snapshot type (route, contact, full)")
	cmd.Flags().StringVar(&note, "note", "", "Snapshot note/description")

	return cmd
}

// newSnapshotListCmd creates the snapshot list command.
func newSnapshotListCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
			defer stateManager.Close()

			snapshots, err := stateManager.ListSnapshots(ctx)
			if err != nil {
				return fmt.Errorf("failed to list snapshots: %w", err)
			}

			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			return outputter.RenderSnapshots(snapshots)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}

// newSnapshotShowCmd creates the snapshot show command.
func newSnapshotShowCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show <snapshot-id>",
		Short: "Show snapshot details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			snapshotID := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
			defer stateManager.Close()

			snapshot, err := stateManager.LoadSnapshot(ctx, snapshotID)
			if err != nil {
				return fmt.Errorf("failed to load snapshot: %w", err)
			}

			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			switch outputFormat {
			case "json":
				return outputter.renderJSON(snapshot)
			case "yaml":
				return outputter.renderYAML(snapshot)
			default:
				fmt.Printf("ID: %s\n", snapshot.ID)
				fmt.Printf("Type: %s\n", snapshot.Type)
				fmt.Printf("Timestamp: %s\n", snapshot.Timestamp.Format("2006-01-02 15:04:05"))
				fmt.Printf("Note: %s\n", snapshot.Note)
				fmt.Printf("Checksum: %s\n", snapshot.Checksum)
				if snapshot.Routes != nil {
					fmt.Printf("Routes: %d\n", snapshot.Routes.Count)
				}
				if snapshot.Contacts != nil {
					fmt.Printf("Contacts: %d\n", snapshot.Contacts.Count)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}

// newSnapshotDeleteCmd creates the snapshot delete command.
func newSnapshotDeleteCmd(logger *logrus.Logger) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <snapshot-id>",
		Short: "Delete a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			snapshotID := args[0]

			if !confirm {
				return fmt.Errorf("please confirm deletion with --confirm flag")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
			defer stateManager.Close()

			if err := stateManager.DeleteSnapshot(ctx, snapshotID); err != nil {
				return fmt.Errorf("failed to delete snapshot: %w", err)
			}

			fmt.Printf("Successfully deleted snapshot %s\n", snapshotID)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	return cmd
}
