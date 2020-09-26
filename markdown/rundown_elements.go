package markdown

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

/*
 *
 *   RUNDOWN EXTENSION
 *
 */

type rundownElements struct {
}

// Strikethrough is an extension that allow you to use codeModifier expression like '~~text~~' .
var RundownElements = &rundownElements{}

func (e *rundownElements) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(
		parser.WithInlineParsers(
			util.Prioritized(NewRundownInlineParser(), 2),
		),
		parser.WithBlockParsers(
			util.Prioritized(NewRundownBlockParser(), 1),
		),
		parser.WithASTTransformers(
			util.Prioritized(NewRundownASTTransformer(), 1),
		),
	)
}

type RundownNode interface {
	ast.Node
	GetModifiers() *Modifiers
}

/*
 *
 *   RUNDOWN INLINE NODE
 *
 */

// A Rundown struct represents an inline <rundown> element.
type RundownInline struct {
	ast.BaseInline
	Segments       *text.Segments
	Modifiers      *Modifiers
	MutateContents func(input []byte) []byte
}

func (n *RundownInline) GetModifiers() *Modifiers {
	return n.Modifiers
}

// Inline implements Inline.Inline.
func (n *RundownInline) Inline() {}

// Dump implements Node.Dump.
func (n *RundownInline) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Modifiers": n.Modifiers.String()}, nil)
}

// KindRundown is a NodeKind of the Rundown node.
var KindRundownInline = ast.NewNodeKind("RundownInline")

// Kind implements Node.Kind.
func (n *RundownInline) Kind() ast.NodeKind {
	return KindRundownInline
}

// NewRundown returns a new Rundown node.
func NewRundownInline(mods *Modifiers) *RundownInline {
	return &RundownInline{
		BaseInline: ast.BaseInline{},
		Segments:   text.NewSegments(),
		Modifiers:  mods,
	}
}

/*
 *
 *   DETECTOR
 *
 */

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

/*
 *
 *   RUNDOWN INLINE PARSER
 *
 */

type RundownInlineParser struct {
}

var defaultRundownInlineParser = &RundownInlineParser{}

// NewRundownInlineParser return a new InlineParser that can parse
// inline htmls
func NewRundownInlineParser() parser.InlineParser {
	return defaultRundownInlineParser
}

func (s *RundownInlineParser) Trigger() []byte {
	return []byte{'<'}
}

var rundownInlineStateKey = parser.NewContextKey()
var rundownBottom = parser.NewContextKey()

type rundownInlineState struct {
	ast.BaseInline
	Modifiers *Modifiers
}

func (s *rundownInlineState) Text(source []byte) []byte {
	return []byte{}
}

func (s *rundownInlineState) Dump(source []byte, level int) {
	fmt.Printf("%srundownInlineState: \"%s\"\n", strings.Repeat("    ", level), s.Text(source))
}

var kindRundownInlineState = ast.NewNodeKind("RundownInlineState")

func (s *rundownInlineState) Kind() ast.NodeKind {
	return kindRundownInlineState
}

func newRundownInlineState() *rundownInlineState {
	return &rundownInlineState{
		Modifiers: NewModifiers(),
	}
}

func (s *RundownInlineParser) Parse(parent ast.Node, block text.Reader, pc parser.Context) ast.Node {
	line, _ := block.PeekLine()

	if bytes.HasPrefix(line, []byte("<r")) {
		if match := matchRundownTag(block, pc); match != nil {

			mods := ParseModifiers(string(line[match[2]:match[3]]), "=")

			// Moved into the Console Renderer.
			// // Is the parent a Heading? Remove the space between tag and text.
			// if _, ok := parent.(*ast.Heading); ok {
			// 	if subject, ok := parent.LastChild().(*ast.Text); ok {
			// 		cutSeg := subject.Segment.WithStop(segment.Start)
			// 		trimmedSeg := cutSeg.TrimRightSpace(block.Source())
			// 		subject.Segment = trimmedSeg
			// 	}
			// }

			// node.Segments.Append(segment.WithStop(segment.Start + match[1]))

			if match[4] > -1 { // Short closing tag attached
				block.Advance(match[5]) // Move to the end of the closing tag.
				return NewRundownInline(mods)
			}

			// Bookmark where we are so we can proess delimiters later.
			pc.Set(rundownBottom, pc.LastDelimiter())

			node := newRundownInlineState()
			node.Modifiers.Ingest(mods)

			pc.Set(rundownInlineStateKey, node)

			block.Advance(match[1])

			return node
		}
	} else if bytes.HasPrefix(line, []byte("</r")) {
		if match := rundownTagClose.FindIndex(line); match != nil && match[0] == 0 {
			block.Advance(match[1])
			state, ok := pc.Get(rundownInlineStateKey).(*rundownInlineState)
			if !ok {
				return nil
			}

			pc.Set(rundownInlineStateKey, nil)

			node := NewRundownInline(state.Modifiers)

			// Process contents.
			var bottom ast.Node
			if v := pc.Get(rundownBottom); v != nil {
				bottom = v.(ast.Node)
			}
			pc.Set(rundownBottom, nil)
			parser.ProcessDelimiters(bottom, pc)

			// Move contents into the Rundown node.
			for c := state.NextSibling(); c != nil; {
				next := c.NextSibling()
				parent.RemoveChild(parent, c)
				node.AppendChild(node, c)
				c = next
			}

			// Remove the state element
			state.Parent().RemoveChild(state.Parent(), state)

			return node
		}
	}

	return nil
}

func (s *RundownInlineParser) CloseBlock(parent ast.Node, pc parser.Context) {
	// nothing to do
}

/*
 *
 *   RUNDOWN BLOCK NODE
 *
 */

type RundownBlock struct {
	ast.BaseBlock

	// ClosureLine is a line that closes this html block.
	ClosureLine text.Segment
	Modifiers   *Modifiers
}

func (n *RundownBlock) GetModifiers() *Modifiers {
	return n.Modifiers
}

// IsRaw implements Node.IsRaw.
func (n *RundownBlock) IsRaw() bool {
	return true
}

// HasClosure returns true if this html block has a closure line,
// otherwise false.
func (n *RundownBlock) HasClosure() bool {
	return n.ClosureLine.Start >= 0
}

// Dump implements Node.Dump.
func (n *RundownBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, map[string]string{"Modifiers": n.Modifiers.String()}, nil)
}

// KindRundownBlock is a NodeKind of the RundownBlock node.
var KindRundownBlock = ast.NewNodeKind("RundownBlock")

// Kind implements Node.Kind.
func (n *RundownBlock) Kind() ast.NodeKind {
	return KindRundownBlock
}

// NewRundownBlock returns a new RundownBlock node.
func NewRundownBlock(modifiers *Modifiers) *RundownBlock {
	return &RundownBlock{
		BaseBlock:   ast.BaseBlock{},
		Modifiers:   modifiers,
		ClosureLine: text.NewSegment(-1, -1),
	}
}

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

func (b *rundownBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	var node *RundownBlock
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
			node = NewRundownBlock(mods)
		}
	}

	if node != nil {
		reader.Advance(segment.Len() - 1)
		node.Lines().Append(segment)
		return node, parser.NoChildren
	}
	return nil, parser.NoChildren
}

func (b *rundownBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *rundownBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// nothing to do
}

func (b *rundownBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *rundownBlockParser) CanAcceptIndentedLine() bool {
	return false
}

/*
 *
 * FencedCodeBlock + Rundown AST Transformer
 *
 */

type rundownASTTransformer struct {
}

var defaultRundownASTTransformer = &rundownASTTransformer{}

// NewFootnoteASTTransformer returns a new parser.ASTTransformer that
// insert a footnote list to the last of the document.
func NewRundownASTTransformer() parser.ASTTransformer {
	return defaultRundownASTTransformer
}

func (a *rundownASTTransformer) Transform(doc *ast.Document, reader text.Reader, pc parser.Context) {
	// Finds FencedCodeBlocks, and transforms their syntax line additions into RundownBlock elements
	// which provides consistency later

	ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if fcb, ok := node.(*ast.FencedCodeBlock); ok && !entering {
			var infoText string = ""

			info := node.(*ast.FencedCodeBlock).Info

			if info != nil {
				infoText = string(info.Text(reader.Source()))
				splitInfo := ""

				if split := strings.SplitN(infoText, " ", 2); len(split) == 2 { // Trim the syntax specifier
					splitInfo = split[1]
				}

				fencedMods := ParseModifiers(splitInfo, ":") // Fenced modifiers separate KV's with :

				rundown := NewRundownBlock(fencedMods)

				fcb.Parent().InsertBefore(fcb.Parent(), fcb, rundown)
			}
			return ast.WalkSkipChildren, nil
		}

		return ast.WalkContinue, nil
	})
}
