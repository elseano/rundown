package ast

import (
	goldast "github.com/yuin/goldmark/ast"
	"golang.org/x/net/html"
	"gopkg.in/guregu/null.v4"
)

type RundownBlock struct {
	goldast.BaseBlock
	TagName string
	Attrs   []html.Attribute
}

// Dump implements Node.Dump.
func (n *RundownBlock) Dump(source []byte, level int) {
	data := map[string]string{}
	for _, a := range n.Attrs {
		data[a.Key] = a.Val
	}

	goldast.DumpHelper(n, source, level, data, nil)
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

func (r *RundownBlock) HasAttr(names ...string) bool {
	for _, name := range names {
		for _, a := range r.Attrs {
			if a.Key == name {
				return true
			}
		}
	}

	return false
}

func (r *RundownBlock) GetAttr(name string) null.String {
	for _, a := range r.Attrs {
		if a.Key == name {
			return null.StringFromPtr(&a.Val)
		}
	}

	return null.StringFromPtr(nil)
}
