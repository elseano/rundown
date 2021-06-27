package parser

import (
	"bytes"
	"regexp"

	"github.com/elseano/rundown/ast"

	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

/*
 *
 *   RUNDOWN BLOCK PARSER
 *
 */

type rundownBlockParser struct {
}

var defaultRundownBlockParser = &rundownBlockParser{}

func NewRundownBlockParser() parser.BlockParser {
	return defaultRundownBlockParser
}

func (b *rundownBlockParser) Trigger() []byte {
	return []byte{'<'}
}

func (b *rundownBlockParser) Open(parent goldast.Node, reader text.Reader, pc parser.Context) (goldast.Node, parser.State) {
	var node *ast.RundownBlock
	line, segment := reader.PeekLine()

	if pos := pc.BlockOffset(); pos < 0 || line[pos] != '<' {
		return nil, parser.NoChildren
	}

	// Check to see if block has no content
	noContentBlock := bytes.HasSuffix(line, []byte("/>\n")) ||
		bytes.HasSuffix(line, []byte("></r>")) ||
		bytes.HasSuffix(line, []byte("></rundown>"))

	// If this block has content, then return nil. This will turn it into a paragraph,
	// and let the RundownInlineParser pick it up.
	if !noContentBlock {
		return nil, parser.NoChildren
	}

	if bytes.HasPrefix(line, []byte("<r")) {
		if match := matchRundownTag(reader, pc); match != nil {
			mods := ParseModifiers(string(line[match[2]:match[3]]), "=")
			node = ast.NewRundownBlock(mods)
		}
	}

	if node != nil {
		reader.Advance(segment.Len() - 1)
		node.Lines().Append(segment)
		return node, parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *rundownBlockParser) Continue(node goldast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *rundownBlockParser) Close(node goldast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *rundownBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *rundownBlockParser) CanAcceptIndentedLine() bool {
	return false
}

var attributePattern = regexp.MustCompile(`(?:\s+[a-zA-Z_:][a-zA-Z0-9:._-]*(?:\s*=\s*(?:[^\"'=<>` + "`" + `\x00-\x20]+|'[^']*'|"[^"]*"))?)`)
var rundownTagRegexp = regexp.MustCompile(`^(?:<(r|rundown))\s*(.*?)\s*(?:(/>)|>)\s*`)

var rundownTagClose = regexp.MustCompile(`</r>|</rundown>`)

func matchRundownTag(block text.Reader, pc parser.Context) []int {
	line, _ := block.PeekLine()
	match := rundownTagRegexp.FindSubmatchIndex(line)

	if match == nil || len(match) == 0 {
		return nil
	}

	return []int{match[0], match[1], match[4], match[5], match[6], match[7]}
}
