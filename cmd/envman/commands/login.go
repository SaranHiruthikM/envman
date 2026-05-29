package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to envman",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		var email, password string

		fmt.Print("Email: ")
		fmt.Scanln(&email)

		fmt.Print("Password: ")
		fmt.Scanln(&password)

		client := cli.NewClient(serverURL)
		token, err := client.Login(email, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		if err := cli.SaveGlobalConfig(token, serverURL); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("✓ Logged in successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("server", "http://localhost:8080", "envman API server URL")
}
