package parser

import (
	"bytes"
	"fmt"
	"regexp"

	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

/** EXTENSION **/

type invisibleBlocks struct {
}

// Strikethrough is an extension that allow you to use invisibleBlock expression like '~~text~~' .
var InvisibleBlocks = &invisibleBlocks{}

func (e *invisibleBlocks) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithBlockParsers(
			util.Prioritized(NewInvisibleBlockParser(m), 100),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewInvisibleBlockASTTransformer(), 100),
		),
	)
}

/** HOLDING BLOCK **/

type invisibleBlockMarker struct {
	ast.BaseBlock
}

func (b *invisibleBlockMarker) Dump(source []byte, level int) {
	ast.DumpHelper(b, source, level, nil, nil)
}

var kindInvisibleBlockMarker = ast.NewNodeKind("InvisibleBlockMarker")

func (b *invisibleBlockMarker) Kind() ast.NodeKind {
	return kindInvisibleBlockMarker
}

/** PARSER **/

var htmlBlockType2OpenRegexp = regexp.MustCompile(`^[ ]{0,3}<!\-\-\-?\~`)
var htmlBlockType2Close = []byte{'-', '-', '>'}

type invisibleBlockParser struct {
	markdown goldmark.Markdown
}

var defaultInvisibleBlockParser = &invisibleBlockParser{}

// NewInvisibleBlockParser returns a new BlockParser that
// parses fenced code blocks.
func NewInvisibleBlockParser(m goldmark.Markdown) parser.BlockParser {
	return &invisibleBlockParser{
		markdown: m,
	}
}

var invisibleBlockInfoKey = parser.NewContextKey()

func (b *invisibleBlockParser) Trigger() []byte {
	return []byte{'<', '-'}
}

func (b *invisibleBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if pos := pc.BlockOffset(); pos < 0 || (line[pos] != '<' && line[pos] != '-') {
		return nil, parser.NoChildren
	}

	rdutil.Logger.Debug().Msgf("In invisible block maybe: %s", string(line))

	// Invisible Blocks are truely invisible in the AST. We just skip the opening and closing.

	if start := htmlBlockType2OpenRegexp.FindIndex(line); start != nil && !bytes.Contains(line, htmlBlockType2Close) {
		rdutil.Logger.Debug().Msgf("In invisible block on multiple lines.")

		// reader.AdvanceLine()
		pc.Set(invisibleBlockInfoKey, true)
		return &invisibleBlockMarker{}, parser.Close
	}

	if bytes.Contains(line, htmlBlockType2Close) && pc.Get(invisibleBlockInfoKey) == true {
		rdutil.Logger.Debug().Msgf("In end of invisible block.")

		// reader.AdvanceLine()

		line, _ := reader.PeekLine()
		text := string(line)
		fmt.Sprintln(text)

		// // Skip past any trailing blank lines, as these will break the block parser loop.
		// for {
		// 	_, segment := reader.PeekLine()
		// 	segment2 := segment.TrimLeftSpace(reader.Source())
		// 	if segment2.IsEmpty() {
		// 		reader.Advance(segment.Len())
		// 	} else {
		// 		break
		// 	}
		// }

		pc.Set(invisibleBlockInfoKey, nil)
		return &invisibleBlockMarker{}, parser.Close
	}

	return nil, parser.NoChildren
}

func (b *invisibleBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *invisibleBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (b *invisibleBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *invisibleBlockParser) CanAcceptIndentedLine() bool {
	return false
}

/*
 *
 * FencedCodeBlock + Rundown AST Transformer
 *
 */

type invisibleBlockASTTransformer struct {
}

var defaultInvisibleBlockASTTransformer = &invisibleBlockASTTransformer{}

// NewFootnoteASTTransformer returns a new parser.ASTTransformer that
// insert a footnote list to the last of the document.
func NewInvisibleBlockASTTransformer() parser.ASTTransformer {
	return defaultInvisibleBlockASTTransformer
}

func (a *invisibleBlockASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	a.doTransform(doc, reader, pc)
}

func (a *invisibleBlockASTTransformer) doTransform(node ast.Node, reader text.Reader, pc parser.Context) {
	// Finds InvisibleBlocks and removes them.
	ib, ok := node.(*invisibleBlockMarker)
	if ok {
		node.Parent().RemoveChild(node.Parent(), ib)
		return
	}

	for child := node.FirstChild(); child != nil; {
		nextChild := child.NextSibling()
		a.doTransform(child, reader, pc)
		child = nextChild
	}
}
