package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/olekukonko/tablewriter"
)

func RenderShortCodes() error {
	util.Debugf("Rendering shortcodes for %s\n", rundownFile)
	rd, err := rundown.LoadFile(rundownFile)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	shortCodes := rd.GetShortCodes()
	if len(shortCodes.Codes) == 0 && len(shortCodes.Options) == 0 {
		return errors.New("No shortcodes in document")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetBorder(false)

	width := util.GetConsoleWidth() - 10

	codeWidth := 20
	for _, code := range shortCodes.Codes {
		codeLen := len(code.Code)
		if flagDefault == code.Code {
			codeLen = codeLen + len(" (default)")
		}

		if codeLen > codeWidth {
			codeWidth = codeLen
		}
	}

	if codeWidth > (width/2)-1 {
		codeWidth = (width / 2) - 1
	}

	descWidth := width - codeWidth

	if descWidth > 20 {
		table.SetColMinWidth(0, codeWidth)
		// table.SetColMinWidth(1, descWidth)
		table.SetColWidth(descWidth)
	}

	table.SetAutoWrapText(true)
	table.SetReflowDuringAutoWrap(true)

	// table.SetColWidth(20)

	// Document Options
	sortedOptions := sort.StringSlice{}

	for k := range shortCodes.Options {
		sortedOptions = append(sortedOptions, k)
	}

	sortedOptions.Sort()

	fmt.Printf("\nSupported shortcodes:\n\n")

	if sortedOptions.Len() > 0 {

		// table.Append([]string{cleanFilename, "The following options are supported"})

		for _, optCode := range sortedOptions {
			opt := shortCodes.Options[optCode]
			spec := ""

			if opt.Default != "" {
				spec = spec + " (default: " + opt.Default + ")"
			} else if opt.Required {
				spec = spec + " (required)"
			}

			table.Append([]string{"+" + opt.Code + "=[" + opt.Type + "]", opt.Description + spec})
		}
	}

	// Document Shortcodes & Options
	list := sort.StringSlice{}

	for _, code := range shortCodes.Codes {
		list = append(list, code.Code)
	}

	list.Sort()

	if len(list) > 0 && len(sortedOptions) > 0 {
		// table.Append([]string{"", ""})
	}

	for _, codeName := range list {
		if codeName == "rundown:help" {
			continue
		}

		code := shortCodes.Codes[codeName]

		display := code.Name

		if code.Description != "" {
			display = code.Description
		}

		if flagDefault == codeName {
			// codeName = codeName + " (" + aurora.Underline("default").String() + ")"
			codeName = codeName + " (default)"
		}

		table.Append([]string{codeName, display})

		sortedOptions := sort.StringSlice{}

		for k := range code.Options {
			sortedOptions = append(sortedOptions, k)
		}

		sortedOptions.Sort()

		if sortedOptions.Len() > 0 {
			// table.Append([]string{"", aurora.Underline("Options").String()})

			for _, optCode := range sortedOptions {
				opt := code.Options[optCode]
				spec := ""

				if opt.Default != "" {
					spec = spec + " (default: " + opt.Default + ")"
				} else if opt.Required {
					spec = spec + " (required)"
				}

				table.Append([]string{"  +" + opt.Code + "=[" + opt.Type + "]", "  " + opt.Description + spec})
			}
		}

	}

	table.Render()
	fmt.Printf("\n")

	return nil
}
