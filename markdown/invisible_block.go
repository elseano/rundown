// This parser should add a Modifier block element to the AST.

package markdown

import (
	"bytes"
	"regexp"

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
			util.Prioritized(NewInvisibleBlockParser(m), 1),
		),
		// parser.WithASTTransformers(
		// 	util.Prioritized(NewInvisibleASTTransformer(), 999),
		// ),

	)
}

/** BLOCK **/

type Invisible struct {
}

func NewInvisible() *Invisible {
	return &Invisible{}
}

type InvisibleBlock struct {
	ast.BaseBlock

	// ClosureLine is a line that closes this html block.
	ClosureLine text.Segment

	Markdown goldmark.Markdown
}

// KindHTMLBlock is a NodeKind of the HTMLBlock node.
var KindInvisibleBlock = ast.NewNodeKind("InvisibleBlock")

// Kind implements Node.Kind.
func (n *InvisibleBlock) Kind() ast.NodeKind {
	return KindInvisibleBlock
}

func NewInvisibleBlock(markdown goldmark.Markdown) *InvisibleBlock {
	return &InvisibleBlock{
		BaseBlock:   ast.BaseBlock{},
		ClosureLine: text.NewSegment(-1, -1),
		Markdown: markdown,
	}
}

func (n *InvisibleBlock) IsRaw() bool { return true }

func (n *InvisibleBlock) HasClosure() bool {
	return true
}

func (n *InvisibleBlock) Dump(source []byte, level int) {
	ast.DumpHelper(n, source, level, nil, nil)
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
	return []byte{'<'}
}

func (b *invisibleBlockParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, segment := reader.PeekLine()
	if pos := pc.BlockOffset(); pos < 0 || line[pos] != '<' {
		return nil, parser.NoChildren
	}

	if htmlBlockType2OpenRegexp.Match(line) && !bytes.Contains(line, htmlBlockType2Close) {
		node := NewInvisibleBlock(b.markdown)
		reader.Advance(segment.Len() - 1)
		return node, parser.NoChildren
	}

	return nil, parser.NoChildren
}

func (b *invisibleBlockParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	htmlBlock := node.(*InvisibleBlock)
	lines := htmlBlock.Lines()
	line, segment := reader.PeekLine()

	if lines.Len() == 1 {
		firstLine := lines.At(0)
		if bytes.Contains(firstLine.Value(reader.Source()), htmlBlockType2Close) {
			return parser.Close
		}
	}
	if bytes.Contains(line, htmlBlockType2Close) {
		htmlBlock.ClosureLine = segment
		reader.Advance(segment.Len())
		return parser.Close
	}

	node.Lines().Append(segment)
	reader.Advance(segment.Len() - 1)

	return parser.Continue | parser.NoChildren
}

func (b *invisibleBlockParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {
	// var buf bytes.Buffer

	// for i := 0; i < node.Lines().Len(); i++ {
	// 	line := node.Lines().At(i)
	// 	buf.Write(line.Value(reader.Source()))
	// }

	// contentReader := text.NewReader(buf.Bytes())
	// parser := b.markdown.Parser()

	// contentDoc := parser.Parse(contentReader)
	// contentDoc.Dump(buf.Bytes(), 0)

	// for child := contentDoc.FirstChild(); child != nil; child = child.NextSibling() {
	// 	node.InsertAfter(node, node, child)
	// }

}

func (b *invisibleBlockParser) CanInterruptParagraph() bool {
	return true
}

func (b *invisibleBlockParser) CanAcceptIndentedLine() bool {
	return false
}

/** AST TRansformer **

type invisibleASTTransformer struct {
}

var defaultInvisibleASTTransformer = &invisibleASTTransformer{}

// NewInvisibleASTTransformer returns a new parser.ASTTransformer that
// insert a invisible list to the last of the document.
func NewInvisibleASTTransformer() parser.ASTTransformer {
	return defaultInvisibleASTTransformer
}

func (a *invisibleASTTransformer) Transform(node *ast.Document, reader text.Reader, pc parser.Context) {
	for child := node.FirstChild(); child != nil;  {
		if invisible, ok := child.(*InvisibleBlock); ok {
			fmt.Printf("Found invisible block %v\n", child.Kind())

			subLines := captureLines(invisible, reader.Source())

			fmt.Println(subLines)

			subReader := text.NewReader([]byte(subLines))
			subDoc := invisible.Parse(subReader)

			subDoc.Dump(subReader.Source(), 1)


			parent := child.Parent()
			
			for subNode := subDoc.FirstChild(); subNode != nil; {
				fmt.Printf("Inserting child node %v\n", subNode.Kind())

				nextSubNode := subNode.NextSibling()
				parent.InsertAfter(parent, child, subNode)
				subNode = nextSubNode
			}

			parent.RemoveChild(parent, child)
		}

		child = child.NextSibling()
	}
}
**/