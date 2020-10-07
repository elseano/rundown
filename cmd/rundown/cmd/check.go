package cmd

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"

	"github.com/kyokomi/emoji"

	gast "github.com/yuin/goldmark/ast"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/yuin/goldmark/text"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [FILENAME]",
	Short: "Checks the Rundown file for errors",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least a filename")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argFilename = args[0]
	},
	Run: checkExec,
}

func checkExec(cmd *cobra.Command, args []string) {

	table := tablewriter.NewWriter(os.Stdout)

	table.SetColumnAlignment([]int{tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
	table.SetCenterSeparator("")
	table.SetColumnSeparator("")
	table.SetRowSeparator("")
	table.SetRowLine(false)
	table.SetHeaderLine(false)
	table.SetBorder(false)
	table.SetAutoWrapText(false)

	md := markdown.PrepareMarkdown()
	b, _ := ioutil.ReadFile(argFilename)

	reader := text.NewReader(b)

	doc := md.Parser().Parse(reader)

	existingLabels := map[string]bool{}

	gast.Walk(doc, func(node gast.Node, entering bool) (gast.WalkStatus, error) {
		if !entering {
			return gast.WalkContinue, nil
		}

		if emojiNode, ok := node.(*markdown.EmojiInline); ok {
			// Ensure valid emoji

			if _, ok := emoji.CodeMap()[":"+emojiNode.EmojiCode+":"]; !ok {
				table.Append([]string{"ERROR", "Unknown emoji code", ":" + emojiNode.EmojiCode + ":", util.NodeLines(emojiNode.Parent(), b)})
			}
		}

		if rd, ok := node.(markdown.RundownNode); ok {

			// Ensure labels are unique
			if label, ok := rd.GetModifiers().Values[rundown.LabelParameter]; ok {
				if _, set := existingLabels[label]; set {
					table.Append([]string{"ERROR", "Label already in use", label, util.NodeLines(rd, b)})
				}

				existingLabels[label] = true
			}

			// Find invalid modifiers
			for _, err := range rundown.ValidateModifiers(rd.GetModifiers()) {
				table.Append([]string{"ERROR", "Unknown rundown attribute", err.Error(), util.NodeLines(rd, b)})
			}
		}

		if _, ok := node.Parent().(*gast.ListItem); ok {
			if cb, cbOk := node.(*gast.FencedCodeBlock); cbOk {
				table.Append([]string{"WARNING", "Code block inside list.", "If this is unintended, separate the list item and the code block with <!-- -->", util.NodeLines(cb, b)})
			}
		}

		return gast.WalkContinue, nil
	})

	table.Render()
}
