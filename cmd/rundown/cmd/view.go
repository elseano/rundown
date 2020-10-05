package cmd

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/elseano/rundown/pkg/markdown"

	"github.com/spf13/cobra"
)

var viewCmd = &cobra.Command{
	Use:   "view [FILENAME]",
	Short: "Renders a markdown file to the console, without running it.",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least the filename")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argFilename = args[0]
	},
	Run: view,
}

func view(cmd *cobra.Command, args []string) {
	md := markdown.PrepareMarkdown()

	b, _ := ioutil.ReadFile(argFilename)
	md.Convert(b, os.Stdout)
}
