package ast

import goldast "github.com/yuin/goldmark/ast"

type SkipInstruction struct {
	goldast.BaseBlock
	Target goldast.Node
}

func NewSkipInstruction(target goldast.Node) *SkipInstruction {
	return &SkipInstruction{
		Target: target,
	}
}

var KindSkipInstruction = goldast.NewNodeKind("SkipInstruction")

func (n *SkipInstruction) Kind() goldast.NodeKind {
	return KindSkipInstruction
}

func (n *SkipInstruction) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}
