package commands

import (
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Link this folder to an envman project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run envman login first")
		}

		var name string
		fmt.Print("Project name: ")
		fmt.Scanln(&name)

		client := cli.NewClient(cfg.ApiURL)
		projectID, _, err := client.CreateProject(cfg.Token, name)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		if err := cli.SaveProjectConfig(projectID, "development"); err != nil {
			return fmt.Errorf("failed to save project config: %w", err)
		}

		fmt.Printf("✓ Project '%s' created and linked\n", name)
		fmt.Println("  .envman.json created — commit this to git")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
