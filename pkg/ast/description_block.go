package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type DescriptionBlock struct {
	goldast.BaseBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewDescriptionBlock() *DescriptionBlock {
	return &DescriptionBlock{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindDescriptionBlock = goldast.NewNodeKind("DescriptionBlock")

// Kind implements Node.Kind.
func (n *DescriptionBlock) Kind() goldast.NodeKind {
	return KindDescriptionBlock
}

func (n *DescriptionBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}

// Walks through the top-level nodes in the AST under document.
// Will return the first help node before the sections begin.
func GetRootHelp(doc goldast.Node) *DescriptionBlock {
	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*SectionPointer); ok {
			return nil
		}

		if db, ok := child.(*DescriptionBlock); ok {
			return db
		}
	}

	return nil
}
