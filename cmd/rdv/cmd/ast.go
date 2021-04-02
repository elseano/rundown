package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark/text"
)

var astCmd = &cobra.Command{
	Use:   "ast [filename]",
	Short: "Prints out the Rundown AST",

	Run: func(cmd *cobra.Command, args []string) {
		os.Exit(runAst(args[0]))
	},

	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return errors.New("Must supply a filename.")
		}

		return nil
	},
}

func runAst(filename string) int {

	if !util.FileExists(filename) {
		print("File not found.\n")
		return -1
	}

	md := rundown.PrepareMarkdown()

	b, err := ioutil.ReadFile(filename)

	if err != nil {
		fmt.Errorf("Error: %s", err.Error())
		return 1
	}

	reader := text.NewReader(b)

	doc := md.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	doc.Dump(b, 0)
	return 0
}
