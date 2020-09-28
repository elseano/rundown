// This parser should add a Modifier block element to the AST.

package markdown

import (
	"fmt"
	"regexp"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

/** EXTENSION **/

type codeModifiers struct {
}

// Strikethrough is an extension that allow you to use codeModifier expression like '~~text~~' .
var CodeModifiers = &codeModifiers{}

func (e *codeModifiers) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(NewCodeModifierBlockParser(), 2),
	))
	// m.Renderer().AddOptions(renderer.WithNodeRenderers(
	// 	util.Prioritized(NewStrikethroughHTMLRenderer(), 500),
	// ))
}

/** BLOCK **/

type CodeModifier struct {
	modifiers string
}

func NewCodeModifier(modifiers string) *CodeModifier {
	return &CodeModifier{modifiers: modifiers}
}

type CodeModifierBlockType int

const (
	// HTMLBlockType1 represents type 1 html blocks
	XmlBlockType CodeModifierBlockType = iota + 1
	// HTMLBlockType2 represents type 2 html blocks
	HiddenLinkBlockType
	// HTMLBlockType3 represents type 3 html blocks
	FencedCodeBlockType

	NoCodeBlockType
)

type CodeModifierBlock struct {
	ast.BaseBlock

	// Type is a type of this html block.
	CodeModifierBlockType CodeModifierBlockType

	// ClosureLine is a line that closes this html block.
	ClosureLine text.Segment

	Modifiers string
}

// KindHTMLBlock is a NodeKind of the HTMLBlock node.
var KindCodeModifierBlock = ast.NewNodeKind("CodeModifierBlock")

// Kind implements Node.Kind.
func (n *CodeModifierBlock) Kind() ast.NodeKind {
	return KindCodeModifierBlock
}

func NewCodeModifierBlock(typ CodeModifierBlockType, modifiers string) *CodeModifierBlock {
	return &CodeModifierBlock{
		BaseBlock:             ast.BaseBlock{},
		Modifiers:             modifiers,
		CodeModifierBlockType: typ,
		ClosureLine:           text.NewSegment(-1, -1),
	}
}

func (n *CodeModifierBlock) Dump(source []byte, level int) {
	m := map[string]string{
		"Modifiers": fmt.Sprintf("%v", n.Modifiers),
		"Type":      fmt.Sprintf("%d", n.CodeModifierBlockType),
	}
	ast.DumpHelper(n, source, level, m, nil)
}

/** PARSER **/

type codeModifierBlockParser struct {
}

var defaultCodeModifierBlockParser = &codeModifierBlockParser{}

// NewCodeModifierBlockParser returns a new BlockParser that
// parses fenced code blocks.
func NewCodeModifierBlockParser() parser.BlockParser {
	return defaultCodeModifierBlockParser
}

var codeModifierBlockInfoKey = parser.NewContextKey()

func (b *codeModifierBlockParser) Trigger() []byte {
	return []byte{'~', '`', '<', '['} // Added < and [ for code modifier starts
}

type codeModifierState struct {
	skip bool
}

// var codeModifierModifiers = regexp.MustCompile("(```|~~~)\\s*[a-z0-9_]+\\s+([a-z_\\s]+)")
var xmlCodeModifiers = regexp.MustCompile("<!---?~\\s*([a-z]{3}.*)\\s* -?-->")
var linkCodeModifiers = regexp.MustCompile("\\[\\]\\(\\s*([a-z]{3}.*)\\s*\\)")

func getCodeModifier(reader text.Reader, pc parser.Context) ([]byte, CodeModifierBlockType) {
	line, _ := reader.PeekLine()
	pos := pc.BlockOffset()

	text := line[pos:]

	if match := xmlCodeModifiers.FindSubmatch(text); match != nil {
		return match[1], XmlBlockType
	} else if match := linkCodeModifiers.FindSubmatch(text); match != nil {
		return match[1], HiddenLinkBlockType
	}

	return nil, NoCodeBlockType
}

func (b *codeModifierBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	modifiers, typ := getCodeModifier(reader, pc)

	if typ == NoCodeBlockType {
		pc.Set(codeModifierBlockInfoKey, false)
		return nil, parser.NoChildren
	}

	if typ == FencedCodeBlockType {
		if pc.Get(codeModifierBlockInfoKey) == true {
			pc.Set(codeModifierBlockInfoKey, false)
			return nil, parser.NoChildren
		}
		pc.Set(codeModifierBlockInfoKey, true)
	}

	node := NewCodeModifierBlock(typ, string(modifiers))
	return node, parser.NoChildren
}

func (b *codeModifierBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	return parser.Close
}

func (b *codeModifierBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
}

func (b *codeModifierBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *codeModifierBlockParser) CanAcceptIndentedLine() bool {
	return false
}
