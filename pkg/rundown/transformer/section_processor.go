package transformer

import (
	"github.com/elseano/rundown/pkg/rundown/ast"
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

// SectionProcessor inserts a SectionEndPointer at the end of a section.
// Sections defined on a heading are terminated at the next heading of equal or structurally higher level.
// Sections defined as being wrapped inside a rundown tag are terminated at the end of the rundown tag.
type SectionProcessor struct {
	SectionPointer *ast.SectionPointer
	openingTag     *RundownHtmlTag
}

func (p *SectionProcessor) Begin(openingTag *RundownHtmlTag) {
	p.openingTag = openingTag
}

func (p *SectionProcessor) End(node goldast.Node, openingTag *RundownHtmlTag, treatments *Treatment) bool {
	if _, ok := p.SectionPointer.StartNode.(*goldast.Heading); !ok {
		if openingTag.tag == p.openingTag.tag {
			endNode := &ast.SectionEnd{
				BaseBlock:      goldast.BaseBlock{},
				SectionPointer: p.SectionPointer,
			}

			node.Parent().InsertAfter(node.Parent(), node, endNode)
		}
	}

	return false
}

func (p *SectionProcessor) Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool {
	if startHeading, ok := p.SectionPointer.StartNode.(*goldast.Heading); ok {
		if heading, ok := node.(*goldast.Heading); ok {
			if heading.Level <= startHeading.Level {

				endNode := &ast.SectionEnd{
					BaseBlock:      goldast.BaseBlock{},
					SectionPointer: p.SectionPointer,
				}

				heading.Parent().InsertBefore(heading.Parent(), heading, endNode)

				return true
			}
		}
	}

	return false
}
