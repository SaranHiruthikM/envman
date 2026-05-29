package commands

import (
	"fmt"
	"os"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local .env to cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run envman login first")
		}

		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run envman init first")
		}

		data, err := os.ReadFile(".env")
		if err != nil {
			return fmt.Errorf("no .env file found in current directory")
		}

		encrypted, err := cli.Encrypt(string(data), cfg.Token)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		client := cli.NewClient(cfg.ApiURL)
		if err := client.PushSecrets(cfg.Token, proj.ProjectID, encrypted); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("✓ Secrets pushed successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
