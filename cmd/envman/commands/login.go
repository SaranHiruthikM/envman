package commands

import "github.com/spf13/cobra"

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Login to envman",
	RunE: func(cmd *cobra.Command, args []string) error {
		// your logic here
		return nil
	},
}

func init() {
	rootCmd.AddCommand(loginCmd)
}
