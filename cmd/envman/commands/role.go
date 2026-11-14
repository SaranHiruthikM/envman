package commands

import (
	"fmt"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var roleCmd = &cobra.Command{
	Use:   "role",
	Short: "Change a teammate's role",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run `envman login` first")
		}
		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run `envman init` first")
		}

		userID, _ := cmd.Flags().GetString("user")
		role, _ := cmd.Flags().GetString("role")

		if userID == "" {
			userID, _ = cli.Prompt("User ID: ")
		}
		if role == "" {
			role, _ = cli.Prompt("New role (member/coordinator): ")
		}
		role = strings.ToLower(role)
		if role != "member" && role != "coordinator" {
			return fmt.Errorf("role must be 'member' or 'coordinator'")
		}

		client := cli.NewClient(cfg.ApiURL)
		if err := client.UpdateRole(cfg.Token, proj.ProjectID, userID, role); err != nil {
			return fmt.Errorf("role update failed: %w", err)
		}

		fmt.Println("Role updated to", role+".")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(roleCmd)
	roleCmd.Flags().String("user", "", "user ID to update")
	roleCmd.Flags().String("role", "", "new role (member/coordinator)")
}
