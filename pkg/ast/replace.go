package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type ContentReplace struct {
	goldast.String
}

// NewContentReplace returns a new ContentReplace node. The contents
// is the un-evaluated replacement, typically environment variables,
// but can be any text.
func NewContentReplace(contents string) *ContentReplace {
	return &ContentReplace{
		String: *goldast.NewString([]byte(contents)),
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindContentReplace = goldast.NewNodeKind("ContentReplace")

// Kind implements Node.Kind.
func (n *ContentReplace) Kind() goldast.NodeKind {
	return KindContentReplace
}

func (n *ContentReplace) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"Variable": string(n.String.Value)}, nil)
}
