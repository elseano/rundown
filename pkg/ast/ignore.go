package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type IgnoreBlock struct {
	goldast.BaseBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewIgnoreBlock() *IgnoreBlock {
	return &IgnoreBlock{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindIgnoreBlock = goldast.NewNodeKind("IgnoreBlock")

// Kind implements Node.Kind.
func (n *IgnoreBlock) Kind() goldast.NodeKind {
	return KindIgnoreBlock
}

func (n *IgnoreBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
