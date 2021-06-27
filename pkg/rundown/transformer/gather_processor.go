package transformer

import (
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

// Gathers nodes directly descending from replacingNode and putting them into newNode.
// If newNode is nil, then replacingNode and all children are deleted instead.
func NewGatherProcessor(replacingNode goldast.Node, newNode goldast.Node) *GatherProcessor {
	return &GatherProcessor{
		newBlockNode:  newNode,
		replacingNode: replacingNode,
	}
}

type GatherProcessor struct {
	newBlockNode  goldast.Node
	replacingNode goldast.Node
}

func (p *GatherProcessor) Begin() {}

func (p *GatherProcessor) End(treatments *Treatment) {
	if p.newBlockNode != nil {
		treatments.Replace(p.replacingNode, p.newBlockNode)
	} else {
		treatments.Remove(p.replacingNode)
	}
}

func (p *GatherProcessor) Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) {
	if node.Parent() == p.replacingNode {
		if p.newBlockNode != nil {
			treatments.AppendChild(p.newBlockNode, node)
		}
	}
}
