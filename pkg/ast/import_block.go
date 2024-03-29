package ast

import (
	"fmt"

	goldast "github.com/yuin/goldmark/ast"
)

// The rundown import block expects to wrap a markdown link to the file to import.
//
// For example: <r import="ci">[Our CI scripts](./ci/RUNDOWN.md)</r>
//
// Will import our the ./ci/RUNDOWN.md file, namespacing all commands under the "ci:" prefix.
type ImportBlock struct {
	goldast.BaseBlock

	ImportPrefix string
}

// NewImportBlock returns a new RundownBlock node.
func NewImportBlock() *ImportBlock {
	return &ImportBlock{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindImportBlock = goldast.NewNodeKind("ImportBlock")

// Kind implements Node.Kind.
func (n *ImportBlock) Kind() goldast.NodeKind {
	return KindImportBlock
}

func (n *ImportBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"Prefix": n.ImportPrefix}, nil)
}

func (n *ImportBlock) IngestChildren(doc goldast.Node) {

	ownerDoc := n.OwnerDocument()

	for child := doc.FirstChild(); child != nil; {
		thisChild := child
		child = child.NextSibling()

		if n.ImportPrefix != "" {
			if sp, ok := thisChild.(*SectionPointer); ok {
				sp.SectionName = fmt.Sprintf("%s:%s", n.ImportPrefix, sp.SectionName)
			}
		}

		ownerDoc.AppendChild(ownerDoc, thisChild)
	}
}

func (n *ImportBlock) GetFilename() string {
	var dest string = ""

	goldast.Walk(n, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		if link, ok := node.(*goldast.Link); ok {
			dest = string(link.Destination)
			return goldast.WalkStop, nil
		}

		return goldast.WalkContinue, nil
	})

	return dest
}

func ProcessImportBlocks(doc goldast.Node) []*ImportBlock {
	importDirectives := []*ImportBlock{}

	goldast.Walk(doc, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if entering {
			return goldast.WalkContinue, nil
		}

		if directive, ok := n.(*ImportBlock); ok {
			importDirectives = append(importDirectives, directive)
		}

		return goldast.WalkContinue, nil
	})

	return importDirectives
}
