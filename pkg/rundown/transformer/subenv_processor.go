package transformer

import (
	"regexp"

	"github.com/elseano/rundown/pkg/rundown/ast"
	goldast "github.com/yuin/goldmark/ast"
	goldtext "github.com/yuin/goldmark/text"
)

// SubEnv Processor replaces all mentions of $ENV_VAR with a EnvironmentSubstitution node.
type SubEnvProcessor struct {
	rundownBlock *ast.RundownBlock
}

var EnvMatcher = regexp.MustCompile(`(\$[A-Z0-9_]+)`)

func (p *SubEnvProcessor) Begin(node *ast.RundownBlock) {
	p.rundownBlock = node
}

func (p *SubEnvProcessor) End(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool {
	// if p.openingTag != openingTag {
	// 	return false
	// }

	return true
}

func ConvertTextForSubenv(node goldast.Node, reader goldtext.Reader, treatments *Treatment) {
	switch node := node.(type) {
	case *goldast.Text:
		contents := node.Text(reader.Source())

		found := EnvMatcher.FindIndex(contents)

		if found != nil {
			parent := node.Parent()
			last := node

			if found[0] > 0 {
				before := goldast.NewText()
				before.Segment = goldtext.NewSegment(node.Segment.Start, node.Segment.Start+found[0])
				parent.InsertAfter(parent, node, before)

				last = before
			}

			repl := ast.NewSubEnvInline(string(contents[found[0]:found[1]]))
			parent.InsertAfter(parent, last, repl)

			if found[1] < len(contents) {
				after := goldast.NewText()
				after.Segment = goldtext.NewSegment(node.Segment.Start+found[1], node.Segment.Stop)
				parent.InsertAfter(parent, repl, after)
			}

			treatments.Remove(node)

		}

	}
}

func (p *SubEnvProcessor) Process(node goldast.Node, reader goldtext.Reader, treatments *Treatment) bool {
	ConvertTextForSubenv(node, reader, treatments)

	return false

}
