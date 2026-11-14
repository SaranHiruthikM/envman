package commands

import (
	"encoding/base64"
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Register a new envman account",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		name, err := cli.Prompt("Name: ")
		if err != nil {
			return err
		}
		email, err := cli.Prompt("Email: ")
		if err != nil {
			return err
		}
		password, err := cli.PromptPassword("Password: ")
		if err != nil {
			return err
		}
		if name == "" || email == "" || password == "" {
			return fmt.Errorf("name, email, and password are required")
		}

		// Generate X25519 keypair and encrypt private key with password-derived KEK.
		priv, pub, err := cli.GenerateUserKeypair()
		if err != nil {
			return fmt.Errorf("keypair generation failed: %w", err)
		}
		salt, err := cli.RandomSalt()
		if err != nil {
			return err
		}
		params := cli.DefaultArgon2idParams()
		kek := cli.DeriveKEK(password, salt, params)
		encPriv, nonce, err := cli.EncryptPrivateKey(priv, kek)
		if err != nil {
			return fmt.Errorf("private-key sealing failed: %w", err)
		}

		req := cli.RegisterRequest{
			Email:                    email,
			Name:                     name,
			Password:                 password,
			PublicKey:                base64.StdEncoding.EncodeToString(pub),
			EncryptedPrivateKey:      base64.StdEncoding.EncodeToString(encPriv),
			EncryptedPrivateKeyNonce: base64.StdEncoding.EncodeToString(nonce),
			KDFSalt:                  base64.StdEncoding.EncodeToString(salt),
			KDFTime:                  params.Time,
			KDFMemory:                params.Memory,
			KDFParallelism:           params.Parallelism,
		}

		client := cli.NewClient(serverURL)
		resp, err := client.Register(req)
		if err != nil {
			return fmt.Errorf("registration failed: %w", err)
		}

		cfg := &cli.GlobalConfig{
			Token:      resp.Data.Token,
			ApiURL:     serverURL,
			UserID:     resp.Data.UserID,
			Email:      email,
			PrivateKey: base64.StdEncoding.EncodeToString(priv),
			PublicKey:  base64.StdEncoding.EncodeToString(pub),
		}
		if err := cli.SaveGlobalConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Registered. Welcome,", name+".")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
	registerCmd.Flags().String("server", DefaultServer, "envman API server URL")
}
