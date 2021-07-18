package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/spf13/cobra"
)

func help(cmd *cobra.Command, args []string) {
	rd, err := rundown.LoadFile(rundownFile)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	cleanFilename := filepath.Base(rundownFile)

	fmt.Printf("\nHelp for %s\n\n", cleanFilename)

	d, _ := rd.GetAST()
	doc := d.(*markdown.SectionedDocument)
	for desc := doc.Sections[0].Description.Front(); desc != nil; desc = desc.Next() {
		if text := desc.Value.(*markdown.RundownBlock).GetModifiers().GetValue("desc"); text != nil {
			fmt.Printf("  %s\n", *text)
		}
	}

	codes, err := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
	if err == nil {
		rd.RunCodesWithoutValidation(codes)
	}

	err = RenderShortCodes()
	if err != nil {
		fmt.Printf("Document %s provides no help information %s.\n", rundownFile, err.Error())
	}

	os.Exit(0)
}
