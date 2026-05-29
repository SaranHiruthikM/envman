package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a teammate to this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run envman login first")
		}

		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run envman init first")
		}

		email, _ := cmd.Flags().GetString("email")
		role, _ := cmd.Flags().GetString("role")

		if email == "" {
			fmt.Print("Email: ")
			fmt.Scanln(&email)
		}
		if role == "" {
			fmt.Print("Role (member/coordinator): ")
			fmt.Scanln(&role)
		}

		client := cli.NewClient(cfg.ApiURL)
		if err := client.InviteMember(cfg.Token, proj.ProjectID, email, role); err != nil {
			return fmt.Errorf("invite failed: %w", err)
		}

		fmt.Printf("✓ %s invited as %s\n", email, role)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inviteCmd)
	inviteCmd.Flags().String("email", "", "Email of teammate to invite")
	inviteCmd.Flags().String("role", "", "Role to assign (member/coordinator)")
}
