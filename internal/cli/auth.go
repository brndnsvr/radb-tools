package cli

import (
	"context"
	"fmt"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication",
	Long:  "Login, logout, and check authentication status.",
}

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with RADb API",
	Long:  "Login to the RADb API using username and password.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Prompt for username
		var username string
		if ctx.Config.Credentials.Username != "" {
			username = ctx.Config.Credentials.Username
			fmt.Printf("Username [%s]: ", username)
			var input string
			fmt.Scanln(&input)
			if input != "" {
				username = input
			}
		} else {
			fmt.Print("Username: ")
			fmt.Scanln(&username)
		}

		if username == "" {
			return fmt.Errorf("username is required")
		}

		// Prompt for password
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		fmt.Println()
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password := string(passwordBytes)

		if password == "" {
			return fmt.Errorf("password is required")
		}

		// Attempt login
		ctxTimeout := context.Background()
		if err := ctx.APIClient.Login(ctxTimeout, username, password); err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Store credentials
		if err := ctx.CredMgr.SetPassword(username, password); err != nil {
			ctx.Logger.Warnf("Failed to store credentials: %v", err)
			fmt.Println("Warning: Credentials were not saved securely")
		}

		// Update config with username
		ctx.Config.Credentials.Username = username
		if err := ctx.Config.Save(); err != nil {
			ctx.Logger.Warnf("Failed to save config: %v", err)
		}

		fmt.Printf("Successfully authenticated as %s\n", username)
		return nil
	},
}

var authStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check authentication status",
	Long:  "Display current authentication status and configured username.",
	RunE: func(cmd *cobra.Command, args []string) error {
		if ctx.Config.Credentials.Username == "" {
			fmt.Println("Status: Not authenticated")
			fmt.Println("\nRun 'radb-client auth login' to authenticate")
			return nil
		}

		fmt.Printf("Username: %s\n", ctx.Config.Credentials.Username)

		// Check if password is stored
		_, err := ctx.CredMgr.GetPassword(ctx.Config.Credentials.Username)
		if err != nil {
			fmt.Println("Status: Credentials not found (need to login)")
		} else {
			fmt.Println("Status: Authenticated (credentials stored)")
		}

		return nil
	},
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Logout and clear credentials",
	Long:  "Clear stored credentials from the system keyring or encrypted file.",
	RunE: func(cmd *cobra.Command, args []string) error {
		username := ctx.Config.Credentials.Username
		if username == "" {
			fmt.Println("Not currently logged in")
			return nil
		}

		// Clear credentials
		if err := ctx.CredMgr.DeleteAll(username); err != nil {
			ctx.Logger.Warnf("Failed to delete credentials: %v", err)
		}

		// Logout from API
		ctxTimeout := context.Background()
		if err := ctx.APIClient.Logout(ctxTimeout); err != nil {
			ctx.Logger.Warnf("API logout warning: %v", err)
		}

		// Clear username from config
		ctx.Config.Credentials.Username = ""
		if err := ctx.Config.Save(); err != nil {
			ctx.Logger.Warnf("Failed to save config: %v", err)
		}

		fmt.Printf("Logged out %s\n", username)
		return nil
	},
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authLogoutCmd)
}
