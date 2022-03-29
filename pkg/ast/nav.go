package ast

import (
	gold_ast "github.com/yuin/goldmark/ast"
)

func FindNode(parent gold_ast.Node, finder func(n gold_ast.Node) bool) gold_ast.Node {
	var foundNode gold_ast.Node

	gold_ast.Walk(parent, func(n gold_ast.Node, entering bool) (gold_ast.WalkStatus, error) {
		if !entering {
			return gold_ast.WalkContinue, nil
		}
		if finder(n) {
			foundNode = n
			return gold_ast.WalkStop, nil
		}

		return gold_ast.WalkContinue, nil
	})

	return foundNode
}
