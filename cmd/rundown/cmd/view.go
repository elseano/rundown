package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/charmbracelet/glamour"
	"github.com/elseano/rundown/pkg/util"
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
		rundownFile = findMarkdownFile(flagFilename)
	},
	Run: view,
}

func view(cmd *cobra.Command, args []string) {
	fmt.Printf("Reading %s\n", rundownFile)
	byteData, err := ioutil.ReadFile(rundownFile)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Read %d bytes\n", len(byteData))

	r, _ := glamour.NewTermRenderer(
		// detect background color and pick either the default dark or light theme
		glamour.WithAutoStyle(),
		// wrap output at specific width
		glamour.WithWordWrap(util.GetConsoleWidth()),
	)

	out, err := r.Render(string(byteData))
	if err != nil {
		panic(err)
	}
	fmt.Println(out)
}
