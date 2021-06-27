package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type OnFailure struct {
	goldast.BaseBlock
	FailureMessageRegexp string
}

// NewRundownBlock returns a new RundownBlock node.
func NewOnFailure() *OnFailure {
	return &OnFailure{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindOnFailure = goldast.NewNodeKind("OnFailure")

// Kind implements Node.Kind.
func (n *OnFailure) Kind() goldast.NodeKind {
	return KindOnFailure
}

func (n *OnFailure) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
