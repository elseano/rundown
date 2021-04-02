package cmd

import (
	"errors"
	"os"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/logrusorgru/aurora"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var argEmojiTerm string

var emojiCmd = &cobra.Command{
	Use:   "emoji [TERM]",
	Short: "Searches for an emoji code",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least a term")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argEmojiTerm = args[0]
	},
	Run: emojiExec,
}

func emojiExec(cmd *cobra.Command, args []string) {
	matches := 0
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetColumnAlignment([]int{tablewriter.ALIGN_CENTER, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetNoWhiteSpace(true)

	table.SetColumnSeparator("")

	for k, v := range emoji.CodeMap() {
		if strings.Contains(k, argEmojiTerm) {
			aliases := []string{}

			for _, al := range emoji.RevCodeMap()[v] {
				if al != k {
					aliases = append(aliases, al)
				}
			}

			table.Append([]string{v, k, aurora.Faint(strings.Join(aliases, " ")).String()})
			matches++
		}
	}

	table.Render()
}
