package cli

import (
	"context"
	"fmt"

	"github.com/bss/radb-client/internal/api"
	"github.com/bss/radb-client/internal/config"
	"github.com/bss/radb-client/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// NewContactCmd creates the contact command and its subcommands.
func NewContactCmd(logger *logrus.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "contact",
		Aliases: []string{"c", "contacts"},
		Short:   "Manage contacts",
		Long:    "Create, read, update, and delete contacts in RADb",
	}

	cmd.AddCommand(
		newContactListCmd(logger),
		newContactShowCmd(logger),
		newContactCreateCmd(logger),
		newContactUpdateCmd(logger),
		newContactDeleteCmd(logger),
	)

	return cmd
}

// newContactListCmd creates the contact list command.
func newContactListCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all contacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			contacts, err := client.ListContacts(ctx)
			if err != nil {
				return fmt.Errorf("failed to list contacts: %w", err)
			}

			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			return outputter.RenderContacts(contacts)
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}

// newContactShowCmd creates the contact show command.
func newContactShowCmd(logger *logrus.Logger) *cobra.Command {
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a specific contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			id := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			contact, err := client.GetContact(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get contact: %w", err)
			}

			outputter := NewOutputter(OutputFormat(outputFormat), nil, true)
			switch outputFormat {
			case "json":
				return outputter.renderJSON(contact)
			case "yaml":
				return outputter.renderYAML(contact)
			default:
				fmt.Printf("ID: %s\n", contact.ID)
				fmt.Printf("Name: %s\n", contact.Name)
				fmt.Printf("Email: %s\n", contact.Email)
				fmt.Printf("Role: %s\n", contact.Role)
				if contact.Phone != "" {
					fmt.Printf("Phone: %s\n", contact.Phone)
				}
				if contact.Organization != "" {
					fmt.Printf("Organization: %s\n", contact.Organization)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format (table, json, yaml)")
	return cmd
}

// newContactCreateCmd creates the contact create command.
func newContactCreateCmd(logger *logrus.Logger) *cobra.Command {
	var (
		name    string
		email   string
		role    string
		phone   string
		org     string
		address []string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new contact",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			contact := &models.Contact{
				Name:         name,
				Email:        email,
				Role:         models.ContactRole(role),
				Phone:        phone,
				Organization: org,
				Address:      address,
			}

			if err := contact.Validate(); err != nil {
				return fmt.Errorf("contact validation failed: %w", err)
			}

			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			if err := client.CreateContact(ctx, contact); err != nil {
				return fmt.Errorf("failed to create contact: %w", err)
			}

			fmt.Printf("Successfully created contact %s\n", contact.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Contact name (required)")
	cmd.Flags().StringVar(&email, "email", "", "Contact email (required)")
	cmd.Flags().StringVar(&role, "role", "tech", "Contact role (admin, tech, billing, abuse)")
	cmd.Flags().StringVar(&phone, "phone", "", "Contact phone")
	cmd.Flags().StringVar(&org, "org", "", "Organization")
	cmd.Flags().StringSliceVar(&address, "address", nil, "Address lines")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("email")

	return cmd
}

// newContactUpdateCmd creates the contact update command.
func newContactUpdateCmd(logger *logrus.Logger) *cobra.Command {
	var (
		name  string
		email string
		role  string
		phone string
		org   string
	)

	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an existing contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			id := args[0]

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			contact, err := client.GetContact(ctx, id)
			if err != nil {
				return fmt.Errorf("failed to get contact: %w", err)
			}

			if name != "" {
				contact.Name = name
			}
			if email != "" {
				contact.Email = email
			}
			if role != "" {
				contact.Role = models.ContactRole(role)
			}
			if phone != "" {
				contact.Phone = phone
			}
			if org != "" {
				contact.Organization = org
			}

			if err := client.UpdateContact(ctx, contact); err != nil {
				return fmt.Errorf("failed to update contact: %w", err)
			}

			fmt.Printf("Successfully updated contact %s\n", contact.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Contact name")
	cmd.Flags().StringVar(&email, "email", "", "Contact email")
	cmd.Flags().StringVar(&role, "role", "", "Contact role")
	cmd.Flags().StringVar(&phone, "phone", "", "Contact phone")
	cmd.Flags().StringVar(&org, "org", "", "Organization")

	return cmd
}

// newContactDeleteCmd creates the contact delete command.
func newContactDeleteCmd(logger *logrus.Logger) *cobra.Command {
	var confirm bool

	cmd := &cobra.Command{
		Use:   "delete <id>",
		Short: "Delete a contact",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			id := args[0]

			if !confirm {
				return fmt.Errorf("please confirm deletion with --confirm flag")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			client := api.NewHTTPClient(cfg.API.BaseURL, cfg.API.Source, cfg.API.Timeout, logger)

			creds, err := config.LoadCredentials()
			if err != nil {
				return fmt.Errorf("not authenticated: please run 'radb-client auth login' first")
			}

			if err := client.Login(ctx, creds.Username, creds.Password); err != nil {
				return fmt.Errorf("authentication failed: %w", err)
			}

			if err := client.DeleteContact(ctx, id); err != nil {
				return fmt.Errorf("failed to delete contact: %w", err)
			}

			fmt.Printf("Successfully deleted contact %s\n", id)
			return nil
		},
	}

	cmd.Flags().BoolVar(&confirm, "confirm", false, "Confirm deletion")
	return cmd
}
