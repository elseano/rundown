package cmd

import (
	"errors"
	"os"
	"sort"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
)

func RenderShortCodes() error {
	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	shortCodes := rd.GetShortCodes()
	if len(shortCodes.Codes) == 0 && len(shortCodes.Options) == 0 {
		return errors.New("No shortcodes in document")
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	// Document Options
	sortedOptions := sort.StringSlice{}

	for k := range shortCodes.Options {
		sortedOptions = append(sortedOptions, k)
	}

	sortedOptions.Sort()

	if sortedOptions.Len() > 0 {

		table.Append([]string{aurora.Bold("Document Options").String(), "", ""})

		for _, optCode := range sortedOptions {
			opt := shortCodes.Options[optCode]
			spec := ""

			if opt.Default != "" {
				spec = spec + " (default: " + opt.Default + ")"
			} else if opt.Required {
				spec = spec + " (required)"
			}

			table.Append([]string{"", "+" + opt.Code + "=[" + opt.Type + "]", opt.Description + spec})
		}
	}

	// Document Shortcodes & Options
	list := sort.StringSlice{}

	for _, code := range shortCodes.Codes {
		list = append(list, code.Code)
	}

	list.Sort()

	if len(list) > 0 && len(sortedOptions) > 0 {
		table.Append([]string{"", "", ""})
	}

	for _, codeName := range list {
		if codeName == "rundown:help" {
			continue
		}

		code := shortCodes.Codes[codeName]

		display := aurora.Bold(code.Name).String()
		if code.Description != "" {
			display = display + "\n" + code.Description
		}

		table.Append([]string{aurora.Bold(code.Code).String(), "", display})

		sortedOptions := sort.StringSlice{}

		for k := range code.Options {
			sortedOptions = append(sortedOptions, k)
		}

		sortedOptions.Sort()

		for _, optCode := range sortedOptions {
			opt := code.Options[optCode]
			spec := ""

			if opt.Default != "" {
				spec = spec + " (default: " + opt.Default + ")"
			} else if opt.Required {
				spec = spec + " (required)"
			}

			table.Append([]string{"", "+" + opt.Code + "=[" + opt.Type + "]", opt.Description + spec})
		}

		table.Append([]string{"", "", ""})
	}

	table.Render()

	return nil
}
