package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Clear the stored envman session",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := cli.DeleteGlobalConfig(); err != nil {
			return fmt.Errorf("failed to clear config: %w", err)
		}
		fmt.Println("Logged out.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(logoutCmd)
}
