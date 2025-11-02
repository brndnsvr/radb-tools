package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/bss/radb-client/internal/models"
	"github.com/bss/radb-client/internal/state"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewRouteCmd creates the route command and its subcommands.
func NewRouteCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "route",
		Aliases: []string{"r", "routes"},
		Short:   "Manage route objects",
		Long:    "Create, read, update, and delete route objects in RADb",
	}

	cmd.AddCommand(
		newRouteListCmd(logger),
		newRouteShowCmd(logger),
		newRouteCreateCmd(logger),
		newRouteUpdateCmd(logger),
		newRouteDeleteCmd(logger),
		newRouteDiffCmd(logger),
	)

	return cmd
}

// newRouteListCmd creates the route list command.
func newRouteListCmd(logger *logrus.Logger) *cobra.Command {
	var (
		outputFormat string
		autoSnapshot bool
		prefix       string
		origin       string
		mntBy        string
	)

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all routes",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()

			// Build filters
			filters := make(map[string]string)
			if prefix != "" {
				filters["prefix"] = prefix
			}
			if origin != "" {
				filters["origin"] = origin
			}
			if mntBy != "" {
				filters["mnt-by"] = mntBy
			}

			// List routes using shared API client (already authenticated)
			routes, err := ctx.APIClient.ListRoutes(cmdCtx, filters)
			if err != nil {
				return fmt.Errorf("failed to list routes: %w", err)
			}

			// Auto-snapshot if enabled
			if autoSnapshot {
				stateManager, _ := state.NewFileManager(ctx.Config.StateDir(), logger)
				defer stateManager.Close()

				snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "Auto-snapshot from route list")
				snapshot.Routes = routes
				if err := snapshot.ComputeChecksum(); err != nil {
					logger.Warnf("Failed to compute snapshot checksum: %v", err)
				}

				if err := stateManager.SaveSnapshot(cmdCtx, snapshot); err != nil {
					logger.Warnf("Failed to save auto-snapshot: %v", err)
				} else {
					logger.Infof("Created snapshot: %s", snapshot.ID)
				}
			}

			// Render output
			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			return outputter.RenderRoutes(routes)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	cmd.Flags().BoolVar(&autoSnapshot, "snapshot", true, "Automatically create a snapshot")
	cmd.Flags().StringVar(&prefix, "prefix", "", "Filter by prefix")
	cmd.Flags().StringVar(&origin, "origin", "", "Filter by origin ASN")
	cmd.Flags().StringVar(&mntBy, "mnt-by", "", "Filter by maintainer")

	return cmd
}

// newRouteShowCmd creates the route show command.
func newRouteShowCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show <prefix> <asn>",
		Short: "Show a specific route",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Get route using shared API client (already authenticated)
			route, err := ctx.APIClient.GetRoute(cmdCtx, prefix, asn)
			if err != nil {
				return fmt.Errorf("failed to get route: %w", err)
			}

			// Render output
			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			switch outputFormat {
			case "json":
				return outputter.renderJSON(route)
			case "yaml":
				return outputter.renderYAML(route)
			default:
				// Pretty print for table format
				fmt.Printf("Route: %s\n", route.Route)
				fmt.Printf("Origin: %s\n", route.Origin)
				fmt.Printf("Maintainers: %s\n", strings.Join(route.MntBy, ", "))
				if len(route.Descr) > 0 {
					fmt.Printf("Description: %s\n", strings.Join(route.Descr, "; "))
				}
				if len(route.Remarks) > 0 {
					fmt.Printf("Remarks: %s\n", strings.Join(route.Remarks, "; "))
				}
				fmt.Printf("Source: %s\n", route.Source)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}

// newRouteCreateCmd creates the route create command.
func newRouteCreateCmd(logger *logrus.Logger) *cobra.Command {
	var (
		descr   []string
		mntBy   []string
		remarks []string
	)

	cmd := &cobra.Command{
		Use:   "create <prefix> <asn>",
		Short: "Create a new route",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Ensure ASN has AS prefix
			if !strings.HasPrefix(asn, "AS") {
				asn = "AS" + asn
			}

			// Create route object
			route := &models.RouteObject{
				Route:   prefix,
				Origin:  asn,
				Descr:   descr,
				MntBy:   mntBy,
				Remarks: remarks,
				Source:  ctx.Config.API.Source,
			}

			// Validate
			if err := route.Validate(); err != nil {
				return fmt.Errorf("route validation failed: %w", err)
			}

			// Create route using shared API client (already authenticated)
			if err := ctx.APIClient.CreateRoute(cmdCtx, route); err != nil {
				return fmt.Errorf("failed to create route: %w", err)
			}

			fmt.Printf("Successfully created route %s\n", route.ID())
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&descr, "descr", nil, "Description(s)")
	cmd.Flags().StringSliceVar(&mntBy, "mnt-by", nil, "Maintainer(s) (required)")
	cmd.Flags().StringSliceVar(&remarks, "remarks", nil, "Remarks")
	cmd.MarkFlagRequired("mnt-by")

	return cmd
}

// newRouteUpdateCmd creates the route update command.
func newRouteUpdateCmd(logger *logrus.Logger) *cobra.Command {
	var (
		descr   []string
		mntBy   []string
		remarks []string
	)

	cmd := &cobra.Command{
		Use:   "update <prefix> <asn>",
		Short: "Update an existing route",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Ensure ASN has AS prefix
			if !strings.HasPrefix(asn, "AS") {
				asn = "AS" + asn
			}

			// Get existing route using shared API client (already authenticated)
			route, err := ctx.APIClient.GetRoute(cmdCtx, prefix, asn)
			if err != nil {
				return fmt.Errorf("failed to get route: %w", err)
			}

			// Update fields if provided
			if len(descr) > 0 {
				route.Descr = descr
			}
			if len(mntBy) > 0 {
				route.MntBy = mntBy
			}
			if len(remarks) > 0 {
				route.Remarks = remarks
			}

			// Update route using shared API client
			if err := ctx.APIClient.UpdateRoute(cmdCtx, route); err != nil {
				return fmt.Errorf("failed to update route: %w", err)
			}

			fmt.Printf("Successfully updated route %s\n", route.ID())
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&descr, "descr", nil, "Description(s)")
	cmd.Flags().StringSliceVar(&mntBy, "mnt-by", nil, "Maintainer(s)")
	cmd.Flags().StringSliceVar(&remarks, "remarks", nil, "Remarks")

	return cmd
}

// newRouteDeleteCmd creates the route delete command.
func newRouteDeleteCmd(logger *logrus.Logger) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <prefix> <asn>",
		Short: "Delete a route",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			prefix := args[0]
			asn := args[1]

			if !confirm {
				return fmt.Errorf("please confirm deletion with --confirm flag")
			}

			// Delete route using shared API client (already authenticated)
			if err := ctx.APIClient.DeleteRoute(cmdCtx, prefix, asn); err != nil {
				return fmt.Errorf("failed to delete route: %w", err)
			}

			fmt.Printf("Successfully deleted route %s-%s\n", prefix, asn)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	return cmd
}

// newRouteDiffCmd creates the route diff command.
func newRouteDiffCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "diff <snapshot-id-1> <snapshot-id-2>",
		Short: "Compare two route snapshots",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmdCtx := context.Background()
			snapshot1ID := args[0]
			snapshot2ID := args[1]

			// Create state manager using shared config
			stateManager, _ := state.NewFileManager(ctx.Config.StateDir(), logger)
			defer stateManager.Close()

			// Load snapshots
			snap1, err := stateManager.LoadSnapshot(cmdCtx, snapshot1ID)
			if err != nil {
				return fmt.Errorf("failed to load snapshot %s: %w", snapshot1ID, err)
			}

			snap2, err := stateManager.LoadSnapshot(cmdCtx, snapshot2ID)
			if err != nil {
				return fmt.Errorf("failed to load snapshot %s: %w", snapshot2ID, err)
			}

			// Compute diff
			diff, err := state.ComputeDiff(cmdCtx, snap1, snap2)
			if err != nil {
				return fmt.Errorf("failed to compute diff: %w", err)
			}

			// Render output
			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			return outputter.RenderDiff(diff)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}
