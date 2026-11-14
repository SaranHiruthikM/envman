package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Version is baked at build time via -ldflags.
var Version = "dev"

const DefaultServer = "http://localhost:8080"

var rootCmd = &cobra.Command{
	Use:   "envman",
	Short: "Team .env file sync with end-to-end encryption",
	Long: `envman keeps your team's .env files in sync.

Secrets are encrypted on your machine with a per-project key you and your
teammates share; the server only ever stores ciphertext.

Typical workflow:

  envman register            # create an account
  envman init                # link this folder to a new project
  envman push                # upload .env (encrypted)
  envman invite --email ...  # add a teammate
  envman get                 # on a teammate's machine, pull latest
`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print envman version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("envman", Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
