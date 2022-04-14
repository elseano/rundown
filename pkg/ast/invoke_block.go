package ast

import (
	"fmt"

	"github.com/elseano/rundown/pkg/util"
	goldast "github.com/yuin/goldmark/ast"
)

type InvokeBlock struct {
	goldast.BaseBlock

	Invoke       string
	AsDependency bool
	Args         map[string]string

	PreviousEnv map[string]string
}

// NewRundownBlock returns a new RundownBlock node.
func NewInvokeBlock() *InvokeBlock {
	return &InvokeBlock{
		BaseBlock: goldast.NewParagraph().BaseBlock,
		Args:      map[string]string{},
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindInvokeBlock = goldast.NewNodeKind("InvokeBlock")

// Kind implements Node.Kind.
func (n *InvokeBlock) Kind() goldast.NodeKind {
	return KindInvokeBlock
}

func (n *InvokeBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{
		"Invoke":       n.Invoke,
		"AsDependency": fmt.Sprintf("%v", n.AsDependency),
		"Args":         fmt.Sprintf("%+v", n.Args),
	}, nil)
}

// Iterates through each child of the given block, identifying InvokeBlocks and copying the invocation contents into the block.
func FillInvokeBlocks(node goldast.Node, maxRecursion int) error {
	return goldast.Walk(node, func(child goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		if invoke, ok := child.(*InvokeBlock); ok {
			section := FindSectionInDocument(node.OwnerDocument(), invoke.Invoke)

			if section == nil {
				return goldast.WalkStop, fmt.Errorf("cannot find section \"%s\"", invoke.Invoke)
			}

			heading := FindNodeBackwards(invoke, func(n goldast.Node) bool {
				_, isHeading := n.(*goldast.Heading)
				return isHeading
			})

			currentLevel := -1
			if heading != nil {
				heading := heading.(*goldast.Heading)
				currentLevel = heading.Level
			}

			invokeHeadingLevel := -1

			count := 0

			var child2 goldast.Node
			for child2 = section.FirstContentNode(); child2 != nil && child2.Kind() != KindSectionEnd; child2 = child2.NextSibling() {
				copied := CopyNode(child2)

				if heading, ok := copied.(*goldast.Heading); ok {
					if invokeHeadingLevel == -1 {
						invokeHeadingLevel = heading.Level
					}

					// Move the headings to be sub-headings under the call point's current heading level.
					fmt.Printf("Invoke at level %d, coming into heading level %d\n", heading.Level, currentLevel)
					heading.Level = heading.Level - invokeHeadingLevel + currentLevel + 1
				}

				if copied != nil {
					count++
					invoke.AppendChild(invoke, copied)
				}
			}

			// Restore the conditional check if there is one.
			if section.HasIfScript() {
				conditional := NewConditionalStart()
				conditional.SetIfScript(section.GetIfScript())

				invoke.InsertBefore(invoke, invoke.FirstChild(), conditional)
				invoke.AppendChild(invoke, conditional.End)
			}

			util.Logger.Debug().Msgf("Added %d children from invoke block %s\n", count, invoke.Invoke)
		}

		return goldast.WalkContinue, nil
	})

}
