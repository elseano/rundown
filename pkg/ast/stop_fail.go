package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type StopFail struct {
	goldast.BaseBlock
	ConditionalImpl
}

// NewRundownBlock returns a new RundownBlock node.
func NewStopFail() *StopFail {
	return &StopFail{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindStopFail = goldast.NewNodeKind("StopFail")

// Kind implements Node.Kind.
func (n *StopFail) Kind() goldast.NodeKind {
	return KindStopFail
}

func (n *StopFail) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
