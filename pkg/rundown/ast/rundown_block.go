package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

/*
 *
 *   RUNDOWN BLOCK NODE
 *
 */

type RundownBlock struct {
	goldast.BaseBlock
}

// IsRaw implements Node.IsRaw.
func (n *RundownBlock) IsRaw() bool {
	return true
}

// Dump implements Node.Dump.
func (n *RundownBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindRundownBlock = goldast.NewNodeKind("RundownBlock")

// Kind implements Node.Kind.
func (n *RundownBlock) Kind() goldast.NodeKind {
	return KindRundownBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewRundownBlock() *RundownBlock {
	return &RundownBlock{
		BaseBlock: goldast.BaseBlock{},
	}
}
