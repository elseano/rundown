package markdown

import (
	"regexp"
	"strings"

	"github.com/kyokomi/emoji"
	"github.com/yuin/goldmark"
	gast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// A EmojiInline struct represents a emoji of GFM text.
type EmojiInline struct {
	gast.BaseInline
	EmojiCode string
}

// Dump implements Node.Dump.
func (n *EmojiInline) Dump(source []byte, level int) {
	gast.DumpHelper(n, source, level, nil, nil)
}

// KindEmojiInline is a NodeKind of the Emoji node.
var KindEmojiInline = gast.NewNodeKind("Emoji")

// Kind implements Node.Kind.
func (n *EmojiInline) Kind() gast.NodeKind {
	return KindEmojiInline
}

// NewEmojiInline returns a new EmojiInline node.
func NewEmojiInline(code string) *EmojiInline {
	return &EmojiInline{EmojiCode: code}
}

type emojiDelimiterProcessor struct {
}

func (p *emojiDelimiterProcessor) IsDelimiter(b byte) bool {
	return b == ':'
}

func (p *emojiDelimiterProcessor) CanOpenCloser(opener, closer *parser.Delimiter) bool {
	return opener.Char == closer.Char
}

func (p *emojiDelimiterProcessor) OnMatch(consumes int) gast.Node {
	return nil
}

var defaultEmojiDelimiterProcessor = &emojiDelimiterProcessor{}

type emojiParser struct {
}

var defaultEmojiParser = &emojiParser{}

// NewEmojiParser return a new InlineParser that parses
// emoji expressions.
func NewEmojiParser() parser.InlineParser {
	return defaultEmojiParser
}

func (s *emojiParser) Trigger() []byte {
	return []byte{':'}
}

var emojiMatch = regexp.MustCompile("\\:([a-z_0-9]+)\\:")

func (s *emojiParser) Parse(parent gast.Node, block text.Reader, pc parser.Context) gast.Node {
	line, _ := block.PeekLine()
	pos := pc.BlockOffset()

	if pos >= 0 && pos < len(line) {
		text := string(line[pos:])

		matches := emojiMatch.FindStringSubmatchIndex(text)
		if matches != nil {
			contents := string(text[matches[2]:matches[3]])
			node := NewEmojiInline(contents)
			block.Advance(len(contents) + 2)
			return node
		}
	}

	return nil

}

func (s *emojiParser) CloseBlock(parent gast.Node, pc parser.Context) {
	// nothing to do
}

// EmojiHTMLRenderer is a renderer.NodeRenderer implementation that
// renders Emoji nodes.
type EmojiRenderer struct {
	html.Config
}

// NewEmojiHTMLRenderer returns a new EmojiHTMLRenderer.
func NewEmojiRenderer() renderer.NodeRenderer {
	r := &EmojiRenderer{}

	return r
}

// RegisterFuncs implements renderer.NodeRenderer.RegisterFuncs.
func (r *EmojiRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindEmojiInline, r.renderEmoji)
}

// EmojiAttributeFilter defines attribute names which dd elements can have.
var EmojiAttributeFilter = html.GlobalAttributeFilter

func (r *EmojiRenderer) renderEmoji(w util.BufWriter, source []byte, n gast.Node, entering bool) (gast.WalkStatus, error) {
	if entering {
		emojiNode := n.(*EmojiInline)
		w.WriteString(strings.TrimSpace(emoji.Sprint(":" + emojiNode.EmojiCode + ":")))
	}
	return gast.WalkContinue, nil
}

type emojiExt struct {
}

// Emoji is an extension that allow you to use emoji expression like '~~text~~' .
var Emoji = &emojiExt{}

func (e *emojiExt) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(NewEmojiParser(), 500),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(NewEmojiRenderer(), 500),
	))
}
