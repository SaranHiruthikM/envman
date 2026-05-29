package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new envman account",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		var email, password, name string

		fmt.Print("Name: ")
		fmt.Scanln(&name)
		fmt.Print("Email: ")
		fmt.Scanln(&email)
		fmt.Print("Password: ")
		fmt.Scanln(&password)

		client := cli.NewClient(serverURL)
		token, err := client.Register(email, password, name)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		if err := cli.SaveGlobalConfig(token, serverURL); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("✓ Registered and logged in successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("server", "http://localhost:8080", "envman API server URL")
}
