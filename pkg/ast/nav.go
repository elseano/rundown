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

// Searches previous siblings, and then previous parent's siblings, and so on
func FindNodeBackwards(currentNode gold_ast.Node, finder func(n gold_ast.Node) bool) gold_ast.Node {
	for node := currentNode.PreviousSibling(); node != nil; node = node.PreviousSibling() {
		if finder(node) {
			return node
		}
	}

	if parent := currentNode.Parent(); parent != nil {
		if finder(parent) {
			return parent
		}

		return FindNodeBackwards(parent, finder)
	}

	return nil
}

// Searches backwards depth-first.
func FindNodeBackwardsDeeply(currentNode gold_ast.Node, finder func(n gold_ast.Node) bool) gold_ast.Node {
	if currentNode == nil {
		return nil
	}

	for node := currentNode.PreviousSibling(); node != nil; node = node.PreviousSibling() {
		if child := FindNodeBackwardsDeeply(node.LastChild(), finder); child != nil {
			return child
		}

		if finder(node) {
			return node
		}

	}

	if parent := currentNode.Parent(); parent != nil {
		if finder(parent) {
			return parent
		}

		return FindNodeBackwardsDeeply(parent.PreviousSibling(), finder)
	}

	return nil
}
