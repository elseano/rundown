package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type SectionGroup struct {
	goldast.BaseBlock

	Name             string
	DescriptionShort string
	DescriptionLong  *DescriptionBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewSectionGroup(name string) *SectionGroup {
	p := &SectionGroup{
		Name: name,
	}

	return p
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindSectionGroup = goldast.NewNodeKind("SectionGroup")

// Kind implements Node.Kind.
func (n *SectionGroup) Kind() goldast.NodeKind {
	return KindSectionGroup
}

func (n *SectionGroup) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"Name": n.Name}, nil)
}
