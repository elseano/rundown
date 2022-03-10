package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type SaveCodeBlock struct {
	goldast.BaseBlock
	CodeBlock *goldast.FencedCodeBlock

	SaveToVariable string
}

// NewRundownBlock returns a new RundownBlock node.
func NewSaveCodeBlock(fcb *goldast.FencedCodeBlock, saveToVariable string) *SaveCodeBlock {
	return &SaveCodeBlock{
		BaseBlock:      goldast.NewParagraph().BaseBlock,
		CodeBlock:      fcb,
		SaveToVariable: saveToVariable,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSaveCodeBlock = goldast.NewNodeKind("SaveCodeBlock")

// Kind implements Node.Kind.
func (n *SaveCodeBlock) Kind() goldast.NodeKind {
	return KindSaveCodeBlock
}

func (n *SaveCodeBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{
		"SaveToVariable": n.SaveToVariable,
	}, nil)
}
