package transformer

import (
	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/util"
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

func PopulateSectionMetadata(start *ast.SectionPointer, reader goldtext.Reader) {
	goldast.Walk(start, func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		util.Logger.Trace().Msgf("Scanning section contents: %s", n.Kind().String())
		switch node := n.(type) {
		case *ast.DescriptionBlock:
			start.DescriptionLong = node
		case *ast.SectionOption:
			start.Options = append(start.Options, node)
		case *ast.InvokeBlock:
			start.Dependencies = append(start.Dependencies, ast.FindSectionInDocument(start.OwnerDocument(), node.Invoke))
		}

		return goldast.WalkContinue, nil
	})
}

// Given the startNode (being the section node itself), returns the last node of the section.
// Called after the section has been inserted into the AST, but before children have been moved into it.
func FindEndOfSection(startNode *goldast.Heading) goldast.Node {
	var lastSib goldast.Node
	for sib := startNode.NextSibling(); sib != nil; sib = sib.NextSibling() {
		util.Logger.Trace().Msgf("Check %s", sib.Kind().String())
		if h, ok := sib.(*goldast.Heading); ok {
			util.Logger.Trace().Msgf("FindEndOfSection: Found heading at %d, section at %d", h.Level, startNode.Level)
			if h.Level <= startNode.Level {
				util.Logger.Trace().Msgf("This section ends here: %T", h)
				return h.PreviousSibling()
			}
		}

		lastSib = sib
	}

	// If we get here, we're the last heading in a section or document.
	return lastSib
}
