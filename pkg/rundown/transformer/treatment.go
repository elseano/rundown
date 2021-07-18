package transformer

import (
	"github.com/elseano/rundown/pkg/util"
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

// Treatment is used to indicate a node should be modified, but batches
// these modifications for after we've walked the AST, otherwise the
// walker gets confused.
type Treatment struct {
	replaceNodes       []func()
	ignoreNodes        map[goldast.Node]bool
	reader             goldtext.Reader
	newNodesStartIndex int
	newNodes           []goldast.Node
}

func NewTreatment(reader goldtext.Reader) *Treatment {
	return &Treatment{
		replaceNodes:       make([]func(), 0),
		ignoreNodes:        map[goldast.Node]bool{},
		reader:             reader,
		newNodesStartIndex: 0,
	}
}

// Returns the newly added nodes since the last call.
func (t *Treatment) NewNodes() []goldast.Node {
	if len(t.newNodes) > 0 {
		index := t.newNodesStartIndex
		t.newNodesStartIndex = len(t.newNodes)
		return t.newNodes[index:]
	}

	return []goldast.Node{}
}

func (t *Treatment) Replace(nodeToReplace goldast.Node, replacement goldast.Node) {
	if replacement != nil {
		t.newNodes = append(t.newNodes, replacement)
	}

	util.Logger.Trace().Msgf("Replacing: %s\n", nodeToReplace.Kind().String())

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

// Deletes a rundown block, moving it's children into it's parent.
func (t *Treatment) DissolveRundownBlock(block goldast.Node) {
	t.replaceNodes = append(t.replaceNodes, func() {
		parent := block.Parent()
		var nextChild goldast.Node
		for child := block.FirstChild(); child != nil; child = nextChild {
			nextChild = child.NextSibling()
			parent.InsertBefore(parent, block, child)
		}

		parent.RemoveChild(parent, block)
	})
}

func (t *Treatment) ReplaceWithChildren(nodeToReplace goldast.Node, replacement goldast.Node, nodeWithChildren goldast.Node) {
	if replacement != nil {
		t.newNodes = append(t.newNodes, replacement)
	}

	t.replaceNodes = append(t.replaceNodes, func() {
		for child := nodeWithChildren.FirstChild(); child != nil; child = child.NextSibling() {
			replacement.AppendChild(replacement, child)
		}

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
	if child != nil {
		t.newNodes = append(t.newNodes, child)
	}

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
