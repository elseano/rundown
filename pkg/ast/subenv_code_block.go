package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type SubEnvBlock struct {
	goldast.BaseBlock

	InnerBlock goldast.Node
}

// NewRundownBlock returns a new RundownBlock node.
func NewSubEnvBlock(innerBlock goldast.Node) *SubEnvBlock {
	return &SubEnvBlock{
		BaseBlock:  goldast.NewParagraph().BaseBlock,
		InnerBlock: innerBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSubEnvBlock = goldast.NewNodeKind("SubEnvBlock")

// Kind implements Node.Kind.
func (n *SubEnvBlock) Kind() goldast.NodeKind {
	return KindSubEnvBlock
}

func (n *SubEnvBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
