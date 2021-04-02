package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rdv [filename] [args]...",
	Short: "Verify a Rundown file",
	Long:  `Parses and verifies a rundown file checking for issues`,
}

func init() {

	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(astCmd)

}

func Execute(version string, gitCommit string) error {
	rootCmd.Version = version + " (" + gitCommit + ")"

	return rootCmd.Execute()
}
