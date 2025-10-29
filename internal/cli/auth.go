package cli

import (
	"context"
	"fmt"
	"os"
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
		fmt.Fprintf(os.Stderr, "[DEBUG] Starting auth login\n")

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
		fmt.Fprintf(os.Stderr, "[DEBUG] Got username: %s\n", username)

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
		fmt.Fprintf(os.Stderr, "[DEBUG] Got password (length: %d)\n", len(password))

		// Attempt login
		fmt.Fprintf(os.Stderr, "[DEBUG] Calling APIClient.Login()\n")
		ctxTimeout := context.Background()
		if err := ctx.APIClient.Login(ctxTimeout, username, password); err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Login failed: %v\n", err)
			return fmt.Errorf("login failed: %w", err)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] APIClient.Login() succeeded\n")

		// Store credentials
		fmt.Fprintf(os.Stderr, "[DEBUG] Storing credentials with CredMgr.SetPassword()\n")
		if err := ctx.CredMgr.SetPassword(username, password); err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SetPassword failed: %v\n", err)
			ctx.Logger.Warnf("Failed to store credentials: %v", err)
			fmt.Println("Warning: Credentials were not saved securely")
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] SetPassword succeeded\n")
		}

		// Update config with username
		fmt.Fprintf(os.Stderr, "[DEBUG] Updating config with username\n")
		ctx.Config.Credentials.Username = username
		if err := ctx.Config.Save(); err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] Config.Save() failed: %v\n", err)
			ctx.Logger.Warnf("Failed to save config: %v", err)
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] Config.Save() succeeded\n")
		}

		fmt.Printf("Successfully authenticated as %s\n", username)
		fmt.Fprintf(os.Stderr, "[DEBUG] Auth login complete\n")
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
