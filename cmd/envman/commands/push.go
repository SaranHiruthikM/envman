package commands

import (
	"encoding/base64"
	"fmt"
	"os"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local .env to the cloud",
	RunE: func(cmd *cobra.Command, args []string) error {
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

		data, err := os.ReadFile(".env")
		if err != nil {
			return fmt.Errorf("no .env file in current directory")
		}

		ciphertext, err := cli.EncryptBody(string(data), projectKey)
		if err != nil {
			return fmt.Errorf("encryption failed: %w", err)
		}

		client := cli.NewClient(cfg.ApiURL)
		if err := client.PushSecrets(cfg.Token, proj.ProjectID, ciphertext); err != nil {
			return fmt.Errorf("push failed: %w", err)
		}

		fmt.Println("Pushed.")
		return nil
	},
}

// loadProjectKey fetches the caller's wrapped project key and unwraps it with
// the locally-stored private key.
func loadProjectKey(cfg *cli.GlobalConfig, proj *cli.ProjectConfig) ([]byte, error) {
	client := cli.NewClient(cfg.ApiURL)
	wrap, err := client.GetSelfKey(cfg.Token, proj.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch project key: %w", err)
	}

	priv, err := base64.StdEncoding.DecodeString(cfg.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("corrupt local private key: %w", err)
	}
	ephPub, err := base64.StdEncoding.DecodeString(wrap.EphemeralPubkey)
	if err != nil {
		return nil, fmt.Errorf("corrupt ephemeral public key: %w", err)
	}
	nonce, err := base64.StdEncoding.DecodeString(wrap.WrappedKeyNonce)
	if err != nil {
		return nil, fmt.Errorf("corrupt nonce: %w", err)
	}
	wrapped, err := base64.StdEncoding.DecodeString(wrap.WrappedKey)
	if err != nil {
		return nil, fmt.Errorf("corrupt wrapped key: %w", err)
	}

	projectKey, err := cli.UnwrapProjectKey(ephPub, nonce, wrapped, priv)
	if err != nil {
		return nil, fmt.Errorf("could not unwrap project key: %w", err)
	}
	return projectKey, nil
}

func init() {
	rootCmd.AddCommand(pushCmd)
}
