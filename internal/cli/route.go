package cli

import (
	"context"
	"fmt"
	"strings"

	"github.com/bss/radb-client/internal/api"
	"github.com/bss/radb-client/internal/config"
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
			ctx := context.Background()

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create API client
			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			// Load credentials and authenticate
			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

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

			// List routes
			routes, err := client.ListRoutes(ctx, filters)
			if err != nil {
				return fmt.Errorf("failed to list routes: %w", err)
			}

			// Auto-snapshot if enabled
			if autoSnapshot {
				stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
				defer stateManager.Close()

				snapshot := models.NewSnapshot(models.SnapshotTypeRoute, "Auto-snapshot from route list")
				snapshot.Routes = routes
				if err := snapshot.ComputeChecksum(); err != nil {
					logger.Warnf("Failed to compute snapshot checksum: %v", err)
				}

				if err := stateManager.SaveSnapshot(ctx, snapshot); err != nil {
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
			ctx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create API client
			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			// Authenticate
			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Get route
			route, err := client.GetRoute(ctx, prefix, asn)
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
			ctx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Ensure ASN has AS prefix
			if !strings.HasPrefix(asn, "AS") {
				asn = "AS" + asn
			}

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create route object
			route := &models.RouteObject{
				Route:   prefix,
				Origin:  asn,
				Descr:   descr,
				MntBy:   mntBy,
				Remarks: remarks,
				Source:  cfg.API.Source,
			}

			// Validate
			if err := route.Validate(); err != nil {
				return fmt.Errorf("route validation failed: %w", err)
			}

			// Create API client
			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			// Authenticate
			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Create route
			if err := client.CreateRoute(ctx, route); err != nil {
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
			ctx := context.Background()
			prefix := args[0]
			asn := args[1]

			// Ensure ASN has AS prefix
			if !strings.HasPrefix(asn, "AS") {
				asn = "AS" + asn
			}

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create API client
			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			// Authenticate
			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Get existing route
			route, err := client.GetRoute(ctx, prefix, asn)
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

			// Update route
			if err := client.UpdateRoute(ctx, route); err != nil {
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
			ctx := context.Background()
			prefix := args[0]
			asn := args[1]

			if !confirm {
				return fmt.Errorf("please confirm deletion with --confirm flag")
			}

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create API client
			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			// Authenticate
			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			// Delete route
			if err := client.DeleteRoute(ctx, prefix, asn); err != nil {
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
			ctx := context.Background()
			snapshot1ID := args[0]
			snapshot2ID := args[1]

			// Load config
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			// Create state manager
			stateManager, _ := state.NewFileManager(cfg.StateDir(), logger)
			defer stateManager.Close()

			// Load snapshots
			snap1, err := stateManager.LoadSnapshot(ctx, snapshot1ID)
			if err != nil {
				return fmt.Errorf("failed to load snapshot %s: %w", snapshot1ID, err)
			}

			snap2, err := stateManager.LoadSnapshot(ctx, snapshot2ID)
			if err != nil {
				return fmt.Errorf("failed to load snapshot %s: %w", snapshot2ID, err)
			}

			// Compute diff
			diff, err := state.ComputeDiff(ctx, snap1, snap2)
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
