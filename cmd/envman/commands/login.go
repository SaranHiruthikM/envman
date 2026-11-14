package commands

import (
	"encoding/base64"
	"fmt"

	"github.com/SaranHiruthikM/envman/internal/cli"
	"github.com/spf13/cobra"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to envman",
	RunE: func(cmd *cobra.Command, args []string) error {
		serverURL, _ := cmd.Flags().GetString("server")

		email, err := cli.Prompt("Email: ")
		if err != nil {
			return err
		}
		password, err := cli.PromptPassword("Password: ")
		if err != nil {
			return err
		}
		if email == "" || password == "" {
			return fmt.Errorf("email and password are required")
		}

		client := cli.NewClient(serverURL)
		resp, err := client.Login(email, password)
		if err != nil {
			return fmt.Errorf("login failed: %w", err)
		}

		// Decrypt the private key with the password-derived KEK.
		salt, err := base64.StdEncoding.DecodeString(resp.Data.KDFSalt)
		if err != nil {
			return fmt.Errorf("corrupt key material: %w", err)
		}
		encPriv, err := base64.StdEncoding.DecodeString(resp.Data.EncryptedPrivateKey)
		if err != nil {
			return fmt.Errorf("corrupt key material: %w", err)
		}
		nonce, err := base64.StdEncoding.DecodeString(resp.Data.EncryptedPrivateKeyNonce)
		if err != nil {
			return fmt.Errorf("corrupt key material: %w", err)
		}

		params := cli.Argon2idParams{
			Time:        resp.Data.KDFTime,
			Memory:      resp.Data.KDFMemory,
			Parallelism: resp.Data.KDFParallelism,
		}
		kek := cli.DeriveKEK(password, salt, params)
		priv, err := cli.DecryptPrivateKey(encPriv, nonce, kek)
		if err != nil {
			return fmt.Errorf("could not decrypt your private key — wrong password?")
		}

		cfg := &cli.GlobalConfig{
			Token:      resp.Data.Token,
			ApiURL:     serverURL,
			UserID:     resp.Data.UserID,
			Email:      email,
			PrivateKey: base64.StdEncoding.EncodeToString(priv),
			PublicKey:  resp.Data.PublicKey,
		}
		if err := cli.SaveGlobalConfig(cfg); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Println("Logged in as", email+".")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
	loginCmd.Flags().String("server", DefaultServer, "envman API server URL")
}
