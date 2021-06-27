package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type SectionPointer struct {
	goldast.BaseInline
	SectionName string
	StartNode   goldast.Node
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionPointer(name string) *SectionPointer {
	return &SectionPointer{
		SectionName: name,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionPointer = goldast.NewNodeKind("SectionPointer")

// Kind implements Node.Kind.
func (n *SectionPointer) Kind() goldast.NodeKind {
	return KindSectionPointer
}

func (n *SectionPointer) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"SectionName": n.SectionName}, nil)
}
