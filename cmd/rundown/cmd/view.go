package cmd

import (
	"fmt"
	"io/ioutil"

	"github.com/charmbracelet/glamour"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

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
