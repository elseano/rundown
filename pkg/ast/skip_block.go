package ast

import (
	"fmt"

	goldast "github.com/yuin/goldmark/ast"
)

type SkipBlock struct {
	goldast.BaseBlock
	ConditionalImpl

	Target goldast.Node
}

func NewSkipBlock() *SkipBlock {
	return &SkipBlock{}
}

var KindSkipBlock = goldast.NewNodeKind("SkipBlock")

func (n *SkipBlock) Kind() goldast.NodeKind {
	return KindSkipBlock
}

func (n *SkipBlock) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{
		"Target": fmt.Sprintf("%T", n.Target),
	}, nil)
}

func (n *SkipBlock) SetTarget(node goldast.Node) {
	n.Target = node
}

func (n *SkipBlock) GetEndSkipNode(goldast.Node) goldast.Node {
	return n.Target
}

func DetermineSkipTarget(n *SkipBlock) goldast.Node {
	mySection := GetSectionForNode(n)

	if mySection == nil {
		return n.OwnerDocument()
	}

	nextSection := GetNextSection(mySection)

	if nextSection == nil {
		return n.OwnerDocument()
	}

	return nextSection
}

// Populates the next AST node for each of the Skip nodes.
func PopulateSkipTargets(doc *goldast.Document) {
	goldast.Walk(doc, func(node goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if skip, ok := node.(*SkipBlock); ok && entering {
			target := DetermineSkipTarget(skip)
			skip.SetTarget(target)
		}
		return goldast.WalkContinue, nil
	})
}
