package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Change a teammate's role",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run envman login first")
		}

		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run envman init first")
		}

		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")

		if userID == "" {
			fmt.Print("User ID: ")
			fmt.Scanln(&userID)
		}
		if role == "" {
			fmt.Print("New role (member/coordinator): ")
			fmt.Scanln(&role)
		}

		client := cli.NewClient(cfg.ApiURL)
		if err := client.UpdateRole(cfg.Token, proj.ProjectID, userID, role); err != nil {
			return fmt.Errorf("role update failed: %w", err)
		}

		fmt.Printf("✓ Role updated to %s\n", role)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(roleCmd)
	roleCmd.Flags().String("user", "", "User ID to update")
	roleCmd.Flags().String("role", "", "New role (member/coordinator)")
}
