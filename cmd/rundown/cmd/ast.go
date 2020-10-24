package cmd

import (
	"errors"
	"io/ioutil"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/yuin/goldmark/text"

	"github.com/spf13/cobra"
)

var astCmd = &cobra.Command{
	Use:   "ast [FILENAME]",
	Short: "Displays the Markdown AST for a rundown/markdown file",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least the filename")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argFilename = args[0]
	},
	Run: ast,
}

func ast(cmd *cobra.Command, args []string) {
	md := rundown.PrepareMarkdown()

	b, _ := ioutil.ReadFile(argFilename)

	reader := text.NewReader(b)

	doc := md.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	doc.Dump(b, 0)

}
