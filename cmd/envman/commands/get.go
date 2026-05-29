package commands

import (
	"fmt"
	"os"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Pull latest .env from cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run envman login first")
		}

		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run envman init first")
		}

		client := cli.NewClient(cfg.ApiURL)
		encrypted, err := client.GetSecrets(cfg.Token, proj.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to get secrets: %w", err)
		}

		decrypted, err := cli.Decrypt(encrypted, cfg.Token)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		_, err = os.Stat(".env")
		if os.IsNotExist(err) {
			os.WriteFile(".env", []byte(decrypted), 0600)
			fmt.Println("✓ .env created")
			return nil
		}

		existing, _ := os.ReadFile(".env")
		diff := cli.Diff(string(existing), decrypted)
		if diff == "" {
			fmt.Println("✓ Already up to date")
			return nil
		}

		fmt.Println(diff)
		fmt.Print("Apply changes? [y/n]: ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" {
			fmt.Println("Aborted")
			return nil
		}

		os.WriteFile(".env", []byte(decrypted), 0600)
		fmt.Println("✓ .env updated")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
}
