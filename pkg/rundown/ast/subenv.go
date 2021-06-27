package ast

import (
	goldast "github.com/yuin/goldmark/ast"
)

type EnvironmentSubstitution struct {
	goldast.String
}

// NewRundownBlock returns a new RundownBlock node.
func NewSubEnvInline(contents string) *EnvironmentSubstitution {
	return &EnvironmentSubstitution{
		String: *goldast.NewString([]byte(contents)),
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindEnvironmentSubstitution = goldast.NewNodeKind("EnvironmentSubstitution")

// Kind implements Node.Kind.
func (n *EnvironmentSubstitution) Kind() goldast.NodeKind {
	return KindEnvironmentSubstitution
}

func (n *EnvironmentSubstitution) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{"Variable": string(n.String.Value)}, nil)
}
