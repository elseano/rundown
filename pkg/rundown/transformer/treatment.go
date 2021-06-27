package transformer

import (
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

// Treatment is used to indicate a node should be modified, but batches
// these modifications for after we've walked the AST, otherwise the
// walker gets confused.
type Treatment struct {
	replaceNodes []func()
	ignoreNodes  map[goldast.Node]bool
	reader       goldtext.Reader
}

func NewTreatment(reader goldtext.Reader) *Treatment {
	return &Treatment{
		replaceNodes: make([]func(), 0),
		ignoreNodes:  map[goldast.Node]bool{},
		reader:       reader,
	}
}

func (t *Treatment) Replace(nodeToReplace goldast.Node, replacement goldast.Node) {
	t.replaceNodes = append(t.replaceNodes, func() {
		parent := nodeToReplace.Parent()
		if parent == nil {
			return // Ignore, already removed.
		}

		if replacement.Parent() == nil {
			nodeToReplace.Parent().ReplaceChild(nodeToReplace.Parent(), nodeToReplace, replacement)
		}
	})
}

// Remove a node. Returns what the next node will be after this node is removed.
func (t *Treatment) Remove(nodeToRemove goldast.Node) goldast.Node {
	t.replaceNodes = append(t.replaceNodes, func() {
		// Handle node already removed
		if nodeToRemove.Parent() != nil {
			// Trim spacing between rundown element and the previous node.
			switch prev := nodeToRemove.PreviousSibling().(type) {
			case *goldast.Text:
				prev.Segment = prev.Segment.TrimRightSpace(t.reader.Source())
			}
			nodeToRemove.Parent().RemoveChild(nodeToRemove.Parent(), nodeToRemove)

		}
	})

	nextNode := nodeToRemove.NextSibling()

	if nodeToRemove.Parent().ChildCount() == 1 {
		if p, ok := nodeToRemove.Parent().(*goldast.Paragraph); ok {
			nextNode = p.NextSibling()
		}
	}

	if nextNode == nil {
		nextNode = nodeToRemove.Parent().NextSibling()
	}

	return nextNode
}

func (t *Treatment) AppendChild(parent goldast.Node, child goldast.Node) {
	t.replaceNodes = append(t.replaceNodes, func() {
		parent.AppendChild(parent, child)
	})
}

func (t *Treatment) Ignore(nodeToIgnore goldast.Node) {
	t.ignoreNodes[nodeToIgnore] = true
}

func (t *Treatment) IsIgnored(nodeInQuestion goldast.Node) bool {
	if ig, ok := t.ignoreNodes[nodeInQuestion]; ok {
		return ig
	}

	return false
}

func (t *Treatment) Process(reader goldtext.Reader) {
	for _, replacement := range t.replaceNodes {
		replacement()
	}

}
