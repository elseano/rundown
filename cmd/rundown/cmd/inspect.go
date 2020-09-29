package cmd

import (
	"errors"
	"fmt"

	"github.com/elseano/rundown/internal/cli"

	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect [FILENAME]",
	Short: "Displays the Rundown AST for a rundown file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least the filename")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argFilename = args[0]
	},
	Run: inspect,
}

func inspect(cmd *cobra.Command, args []string) {
	logger := cli.BuildLogger(flagDebug)

	loadedSegments := cli.FileToSegments(argFilename, logger)

	for _, x := range loadedSegments {
		fmt.Println(x.String())
	}

}
