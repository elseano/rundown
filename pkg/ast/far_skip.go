package ast

import gold_ast "github.com/yuin/goldmark/ast"

type FarSkip interface {
	GetEndSkipNode(gold_ast.Node) gold_ast.Node
}
