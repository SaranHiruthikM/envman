package commands

import (
	"encoding/base64"
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Link this folder to a new envman project",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := cli.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("not logged in — run `envman login` first")
		}

		name, err := cli.Prompt("Project name: ")
		if err != nil {
			return err
		}
		if name == "" {
			return fmt.Errorf("project name is required")
		}

		client := cli.NewClient(cfg.ApiURL)
		projectID, err := client.CreateProject(cfg.Token, name)
		if err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}

		// Generate project key and wrap it for self so future machines can recover it.
		projectKey, err := cli.RandomProjectKey()
		if err != nil {
			return err
		}
		myPub, err := base64.StdEncoding.DecodeString(cfg.PublicKey)
		if err != nil {
			return fmt.Errorf("corrupt local public key: %w", err)
		}
		ephPub, nonce, wrapped, err := cli.WrapProjectKey(projectKey, myPub)
		if err != nil {
			return fmt.Errorf("key wrap failed: %w", err)
		}

		if err := client.SetSelfKey(cfg.Token, projectID, cli.ProjectKeyPayload{
			WrappedKey:      base64.StdEncoding.EncodeToString(wrapped),
			WrappedKeyNonce: base64.StdEncoding.EncodeToString(nonce),
			EphemeralPubkey: base64.StdEncoding.EncodeToString(ephPub),
		}); err != nil {
			return fmt.Errorf("failed to upload project key: %w", err)
		}

		if err := cli.SaveProjectConfig(&cli.ProjectConfig{
			ProjectID:   projectID,
			Environment: "development",
		}); err != nil {
			return fmt.Errorf("failed to save project config: %w", err)
		}

		fmt.Printf("Project %q created and linked.\n", name)
		fmt.Println(".envman.json written — commit this file.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
