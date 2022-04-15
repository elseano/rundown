package ast

import (
	"regexp"

	goldast "github.com/yuin/goldmark/ast"
)

type OnFailure struct {
	goldast.BaseBlock
	FailureMessageRegexp string
}

// NewRundownBlock returns a new RundownBlock node.
func NewOnFailure() *OnFailure {
	return &OnFailure{
		BaseBlock: goldast.NewParagraph().BaseBlock,
	}
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindOnFailure = goldast.NewNodeKind("OnFailure")

// Kind implements Node.Kind.
func (n *OnFailure) Kind() goldast.NodeKind {
	return KindOnFailure
}

func (n *OnFailure) Dump(source []byte, level int) {
	goldast.DumpHelper(n, source, level, map[string]string{}, nil)
}

func (n *OnFailure) MatchesError(output []byte) bool {
	if n.FailureMessageRegexp != "" {
		re := regexp.MustCompile(n.FailureMessageRegexp)
		return re.Match(output)
	} else {
		return true
	}
}

func (n *OnFailure) ConvertToParagraph() *goldast.Paragraph {
	para := goldast.NewParagraph()
	for n.ChildCount() > 0 {
		para.AppendChild(para, n.FirstChild())
	}

	return para
}

// Searches the current node's section for OnFailure nodes and returns them.
func GetOnFailureNodes(node goldast.Node) []*OnFailure {
	section := GetSectionForNode(node)

	if section == nil {
		return []*OnFailure{}
	}

	result := []*OnFailure{}

	sectionReached := false

	goldast.Walk(section.OwnerDocument(), func(n goldast.Node, entering bool) (goldast.WalkStatus, error) {
		if !entering {
			return goldast.WalkContinue, nil
		}

		if sectionReached {
			if failure, ok := n.(*OnFailure); ok {
				result = append(result, failure)
			}

			// if n.Kind() == KindSectionEnd {
			// 	return goldast.WalkStop, nil
			// }
		} else {
			if n == section {
				sectionReached = true
			}
		}

		return goldast.WalkContinue, nil

	})

	return result
}
