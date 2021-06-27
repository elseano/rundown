package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type StopOk struct {
	goldast.BaseBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewStopOk() *StopOk {
	return &StopOk{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindStopOk = goldast.NewNodeKind("StopOk")

// Kind implements Node.Kind.
func (n *StopOk) Kind() goldast.NodeKind {
	return KindStopOk
}

func (n *StopOk) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
