package transformer

import (
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

func Replace(nodeToReplace goldast.Node, replacement goldast.Node) {
	parent := nodeToReplace.Parent()
	if parent == nil {
		return // Ignore, already removed.
	}

	if replacement.Parent() == nil {
		nodeToReplace.Parent().ReplaceChild(nodeToReplace.Parent(), nodeToReplace, replacement)
	}
}

// Deletes a rundown block, moving it's children into it's parent.
func DissolveRundownBlock(block goldast.Node) {
	parent := block.Parent()
	var nextChild goldast.Node
	for child := block.FirstChild(); child != nil; child = nextChild {
		nextChild = child.NextSibling()
		parent.InsertBefore(parent, block, child)
	}

	parent.RemoveChild(parent, block)
}

// Replaces the node, and adds children to it from the given node.
func ReplaceWithChildren(nodeToReplace goldast.Node, replacement goldast.Node, nodeWithChildren goldast.Node) {
	child := nodeWithChildren.FirstChild()

	for nodeWithChildren.HasChildren() {
		nextChild := child.NextSibling()

		replacement.AppendChild(replacement, child)
		child = nextChild
	}

	parent := nodeToReplace.Parent()
	if parent == nil {
		return // Ignore, already removed.
	}

	if replacement.Parent() == nil {
		nodeToReplace.Parent().ReplaceChild(nodeToReplace.Parent(), nodeToReplace, replacement)
	}
}

// Remove a node. Returns what next node after this node is removed.
func Remove(nodeToRemove goldast.Node, reader goldtext.Reader) goldast.Node {
	nextNode := nodeToRemove.NextSibling()

	if nextNode == nil {
		nextNode = nodeToRemove.Parent().NextSibling()
	}

	// Handle node already removed
	if nodeToRemove.Parent() != nil {
		// Trim spacing between rundown element and the previous node.
		switch prev := nodeToRemove.PreviousSibling().(type) {
		case *goldast.Text:
			prev.Segment = prev.Segment.TrimRightSpace(reader.Source())
		}
		nodeToRemove.Parent().RemoveChild(nodeToRemove.Parent(), nodeToRemove)

	}

	return nextNode
}

func AppendChild(parent goldast.Node, child goldast.Node) {
	parent.AppendChild(parent, child)
}
