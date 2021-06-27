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
	openingTag    *RundownHtmlTag
}

func (p *GatherProcessor) Begin(openingTag *RundownHtmlTag) {
	p.openingTag = openingTag
}

func (p *GatherProcessor) End(node goldast.Node, openingTag *RundownHtmlTag, treatments *Treatment) bool {
	if p.openingTag != openingTag {
		return false
	}

	if p.newBlockNode != nil {
		treatments.Replace(p.replacingNode, p.newBlockNode)
	} else {
		treatments.Remove(p.replacingNode)
	}

	return true
}

func (p *GatherProcessor) Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool {
	if node.Parent() == p.replacingNode {
		if p.newBlockNode != nil {
			treatments.AppendChild(p.newBlockNode, node)
		}
	}

	return false
}
