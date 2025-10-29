package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/bss/radb-client/internal/api"
	"github.com/bss/radb-client/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// NewWizardCmd creates the interactive configuration wizard command.
func NewWizardCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive configuration wizard",
		Long:  "Run an interactive wizard to set up your RADb client configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWizard(logger)
		},
	}

	return cmd
}

func runWizard(logger *logrus.Logger) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("RADb Client Configuration Wizard")
	fmt.Println("=================================")
	fmt.Println()
	fmt.Println("This wizard will help you configure your RADb client.")
	fmt.Println()

	// Load existing config or create new one
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	// API Configuration
	fmt.Println("API Configuration")
	fmt.Println("-----------------")

	baseURL := promptWithDefault(reader, "API Base URL", cfg.API.BaseURL)
	cfg.API.BaseURL = baseURL

	source := promptWithDefault(reader, "IRR Source", cfg.API.Source)
	cfg.API.Source = source

	timeout := promptWithDefault(reader, "Timeout (seconds)", fmt.Sprintf("%d", cfg.API.Timeout))
	fmt.Sscanf(timeout, "%d", &cfg.API.Timeout)

	// Credentials
	fmt.Println()
	fmt.Println("Authentication")
	fmt.Println("--------------")

	username := prompt(reader, "Username")
	password := promptPassword("Password")

	// Test connection
	fmt.Println()
	fmt.Println("Testing connection...")

	ctx := context.Background()
	client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

	if err := client.Login(ctx, username, password); err != nil {
		fmt.Printf("Warning: Connection test failed: %v\n", err)
		fmt.Println("Configuration will be saved anyway.")
	} else {
		fmt.Println("Connection test successful!")
	}

	// Save configuration
	fmt.Println()
	fmt.Print("Save configuration? [Y/n]: ")
	confirm, _ := reader.ReadString('\n')
	confirm = strings.TrimSpace(confirm)

	if confirm == "" || strings.ToLower(confirm) == "y" {
		if err := config.Save(cfg); err != nil {
			return fmt.Errorf("failed to save configuration: %w", err)
		}

		// Save credentials
		creds := &config.Credentials{
			Username: username,
			Password: password,
		}
		if err := config.SaveCredentials(creds); err != nil {
			return fmt.Errorf("failed to save credentials: %w", err)
		}

		fmt.Println()
		fmt.Println("Configuration saved successfully!")
		fmt.Println()
		fmt.Println("You can now use the RADb client. Try:")
		fmt.Println("  radb-client route list")
		fmt.Println("  radb-client contact list")
	}

	return nil
}

func prompt(reader *bufio.Reader, question string) string {
	fmt.Printf("%s: ", question)
	answer, _ := reader.ReadString('\n')
	return strings.TrimSpace(answer)
}

func promptWithDefault(reader *bufio.Reader, question, defaultValue string) string {
	fmt.Printf("%s [%s]: ", question, defaultValue)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if answer == "" {
		return defaultValue
	}
	return answer
}

func promptPassword(question string) string {
	fmt.Printf("%s: ", question)
	password, _ := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	return string(password)
}
