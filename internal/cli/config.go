package cli

import (
	"fmt"

	"github.com/bss/radb-client/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
	Long:  "Initialize, view, and modify RADb client configuration.",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration",
	Long:  "Create a new configuration file with default values.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Initialize()
		if err != nil {
			return err
		}

		fmt.Printf("Configuration initialized at: %s\n", cfg.ConfigFile)
		fmt.Printf("Cache directory: %s\n", cfg.Preferences.CacheDir)
		fmt.Printf("History directory: %s\n", cfg.Preferences.HistoryDir)
		fmt.Println("\nNext steps:")
		fmt.Println("1. Run 'radb-client auth login' to authenticate")
		fmt.Println("2. Run 'radb-client config show' to view current configuration")

		return nil
	},
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Long:  "Display the current configuration settings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Printf("Configuration File: %s\n\n", ctx.Config.ConfigFile)
		fmt.Println("API Settings:")
		fmt.Printf("  Base URL: %s\n", ctx.Config.API.BaseURL)
		fmt.Printf("  Source: %s\n", ctx.Config.API.Source)
		fmt.Printf("  Format: %s\n", ctx.Config.API.Format)
		fmt.Printf("  Timeout: %ds\n", ctx.Config.API.Timeout)

		fmt.Println("\nRate Limiting:")
		fmt.Printf("  Requests/min: %d\n", ctx.Config.API.RateLimit.RequestsPerMinute)
		fmt.Printf("  Burst size: %d\n", ctx.Config.API.RateLimit.BurstSize)

		fmt.Println("\nPreferences:")
		fmt.Printf("  Cache dir: %s\n", ctx.Config.Preferences.CacheDir)
		fmt.Printf("  History dir: %s\n", ctx.Config.Preferences.HistoryDir)
		fmt.Printf("  Log level: %s\n", ctx.Config.Preferences.LogLevel)

		fmt.Println("\nCredentials:")
		if ctx.Config.Credentials.Username != "" {
			fmt.Printf("  Username: %s\n", ctx.Config.Credentials.Username)
		} else {
			fmt.Println("  Username: (not configured)")
		}

		return nil
	},
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Long:  "Set a configuration value. Supported keys: api.base_url, api.source, api.timeout, preferences.log_level",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		// Set the value based on key
		switch key {
		case "api.base_url":
			ctx.Config.API.BaseURL = value
		case "api.source":
			ctx.Config.API.Source = value
		case "preferences.log_level":
			ctx.Config.Preferences.LogLevel = value
		default:
			return fmt.Errorf("unsupported configuration key: %s", key)
		}

		// Save configuration
		if err := ctx.Config.Save(); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configSetCmd)
}
