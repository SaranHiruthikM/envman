package commands

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a teammate to this project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run `envman login` first")
		}
		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run `envman init` first")
		}

		email, _ := cmd.Flags().GetString("email")
		role, _ := cmd.Flags().GetString("role")

		if email == "" {
			email, _ = cli.Prompt("Email: ")
		}
		if role == "" {
			role, _ = cli.Prompt("Role (member/coordinator): ")
		}
		if email == "" || role == "" {
			return fmt.Errorf("email and role are required")
		}
		role = strings.ToLower(role)
		if role != "member" && role != "coordinator" {
			return fmt.Errorf("role must be 'member' or 'coordinator'")
		}

		client := cli.NewClient(cfg.ApiURL)

		// Fetch invitee's public key.
		_, pubB64, err := client.GetUserPublicKey(cfg.Token, email)
		if err != nil {
			return fmt.Errorf("could not fetch public key for %s: %w", email, err)
		}
		recipientPub, err := base64.StdEncoding.DecodeString(pubB64)
		if err != nil {
			return fmt.Errorf("corrupt recipient public key: %w", err)
		}

		// Unwrap our own copy of the project key so we can re-wrap for them.
		projectKey, err := loadProjectKey(cfg, proj)
		if err != nil {
			return err
		}

		ephPub, nonce, wrapped, err := cli.WrapProjectKey(projectKey, recipientPub)
		if err != nil {
			return fmt.Errorf("could not wrap project key for invitee: %w", err)
		}

		if err := client.InviteMember(cfg.Token, proj.ProjectID, cli.InviteMemberRequest{
			Email:           email,
			Role:            role,
			WrappedKey:      base64.StdEncoding.EncodeToString(wrapped),
			WrappedKeyNonce: base64.StdEncoding.EncodeToString(nonce),
			EphemeralPubkey: base64.StdEncoding.EncodeToString(ephPub),
		}); err != nil {
			return fmt.Errorf("invite failed: %w", err)
		}

		fmt.Printf("%s invited as %s.\n", email, role)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(inviteCmd)
	inviteCmd.Flags().String("email", "", "email of teammate to invite")
	inviteCmd.Flags().String("role", "", "role to assign (member/coordinator)")
}
