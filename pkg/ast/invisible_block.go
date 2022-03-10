package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type InvisibleBlock struct {
	goldast.BaseBlock

	ImportPrefix string
}

// NewRundownBlock returns a new RundownBlock node.
func NewInvisibleBlock() *InvisibleBlock {
	return &InvisibleBlock{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindInvisibleBlock = goldast.NewNodeKind("InvisibleBlock")

// Kind implements Node.Kind.
func (n *InvisibleBlock) Kind() goldast.NodeKind {
	return KindInvisibleBlock
}

func (n *InvisibleBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"Prefix": n.ImportPrefix}, nil)
}
