package commands

import (
	"fmt"
	"os"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Pull the latest .env from the cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
		yes, _ := cmd.Flags().GetBool("yes")

		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run `envman login` first")
		}
		proj, err := cli.LoadProjectConfig()
		if err != nil {
			return fmt.Errorf("no project linked — run `envman init` first")
		}

		projectKey, err := loadProjectKey(cfg, proj)
		if err != nil {
			return err
		}

		client := cli.NewClient(cfg.ApiURL)
		ciphertext, err := client.GetSecrets(cfg.Token, proj.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to fetch secrets: %w", err)
		}

		decrypted, err := cli.DecryptBody(ciphertext, projectKey)
		if err != nil {
			return fmt.Errorf("decryption failed: %w", err)
		}

		if _, err := os.Stat(".env"); os.IsNotExist(err) {
			if err := os.WriteFile(".env", []byte(decrypted), 0600); err != nil {
				return err
			}
			fmt.Println(".env created.")
			return nil
		}

		existing, _ := os.ReadFile(".env")
		diff := cli.Diff(string(existing), decrypted)
		if diff == "" {
			fmt.Println("Already up to date.")
			return nil
		}

		fmt.Println(diff)

		if !yes {
			if !cli.IsStdinTerminal() {
				return fmt.Errorf("refusing to overwrite non-interactively without --yes")
			}
			if !cli.Confirm("Apply changes? [y/n]: ") {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Back up the existing file before overwriting.
		if err := os.WriteFile(".env.bak", existing, 0600); err != nil {
			return fmt.Errorf("failed to write backup: %w", err)
		}
		if err := os.WriteFile(".env", []byte(decrypted), 0600); err != nil {
			return err
		}
		fmt.Println(".env updated. Previous contents saved to .env.bak.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(getCmd)
	getCmd.Flags().Bool("yes", false, "overwrite .env without prompting (for CI)")
}
