package transformer

import (
	"strings"

	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/util"
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

func (p *SectionProcessor) End(node goldast.Node, reader goldtext.Reader, openingTag *RundownHtmlTag, treatments *Treatment) bool {
	if heading, ok := p.SectionPointer.StartNode.(*goldast.Heading); !ok {
		if openingTag != nil && openingTag.tag == p.openingTag.tag {
			endNode := &ast.SectionEnd{
				BaseBlock:      goldast.BaseBlock{},
				SectionPointer: p.SectionPointer,
			}

			node.Parent().InsertAfter(node.Parent(), node, endNode)
		}
	} else {
		// Otherwise, Process will insert the SectionEnd node.
		contents := heading.Text(reader.Source())
		p.SectionPointer.DescriptionShort = strings.Trim(string(contents), " ") // Text hasn't been trimmed yet.
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

	if option, ok := node.(*ast.SectionOption); ok {
		p.SectionPointer.AddOption(option)
	} else if desc, ok := node.(*ast.DescriptionBlock); ok {
		p.SectionPointer.DescriptionLong = desc
	}

	return false
}

// Given the startNode (being the section node itself), returns the node which terminates the section.
func FindEndOfSection(startNode *goldast.Heading) goldast.Node {
	for sib := startNode.NextSibling(); sib != nil; sib = sib.NextSibling() {
		util.Logger.Trace().Msgf("Check %s\n", sib.Kind().String())
		if h, ok := sib.(*goldast.Heading); ok {
			if h.Level <= startNode.Level {
				return h
			}
		}
	}

	return nil
}

func PopulateSectionMetadata(start *ast.SectionPointer, end *ast.SectionEnd, reader goldtext.Reader) {
	inside := false
	goldast.Walk(start.Parent(), func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		if inside {
			util.Logger.Trace().Msgf("Scanning section contents: %s\n", n.Kind().String())
			switch node := n.(type) {
			case *ast.DescriptionBlock:
				start.DescriptionLong = node
			case *ast.SectionOption:
				start.Options = append(start.Options, node)
			}
		}

		if n == start {
			inside = true
		}

		if n == end {
			return goldast.WalkStop, nil
		}

		return goldast.WalkContinue, nil
	})
}
