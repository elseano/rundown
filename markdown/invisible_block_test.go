package markdown

import (
	"testing"

	"github.com/elseano/rundown/util"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestInvisibleBlock1(t *testing.T) {
	contents := []byte("Normal markdown text\n\n<!--~\nHidden markdown text, only for rundown\n-->\nMore text")
	verifyParsedAsKinds(contents, t, []ast.NodeKind{ast.KindParagraph, KindInvisibleBlock, ast.KindParagraph})
}

func TestInvisibleBlock2(t *testing.T) {
	contents := []byte("Normal markdown text\n\n<!--~\nHidden markdown text, only for rundown\n\n    Indented code block\n\n```\nSomeCode\n\n```\n-->\nHere's more normal text")
	verifyParsedAsKinds(contents, t, []ast.NodeKind{ast.KindParagraph, KindInvisibleBlock, ast.KindParagraph})
}

func TestInvisibleBlock3(t *testing.T) {
	contents := []byte("Normal markdown text\n\n<!--~\nHidden markdown text, only for rundown\n-->")
	verifyParsedAsKinds(contents, t, []ast.NodeKind{ast.KindParagraph, KindInvisibleBlock})
}

func TestInvisibleBlock4(t *testing.T) {
	contents := []byte("<!--~ nospin -->\n\n```\nSomeCode\n\n```\nHere's more normal text")
	verifyParsedAsKinds(contents, t, []ast.NodeKind{KindCodeModifierBlock, ast.KindFencedCodeBlock, ast.KindParagraph})
}

func verifyParsedAsKinds(contents []byte, t *testing.T, expected []ast.NodeKind) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			InvisibleBlocks,
			CodeModifiers,
		),
	)

	t.Log(string(contents))

	reader := text.NewReader([]byte(contents))
	doc := markdown.Parser().Parse(reader)

	dump := util.CaptureStdout(func() { doc.Dump(contents, 0) })
	t.Log(dump)

	var child ast.Node
	var i int

	for child, i = doc.FirstChild(), 0; i < len(expected); child, i = child.NextSibling(), i+1 {
		kind := "nil"
		if child != nil {
			kind = child.Kind().String()
		}
		if kind != expected[i].String() {
			t.Fatalf("Expected %v, got %v", expected[i], kind)
		}
	}

	if child != nil {
		t.Fatalf("Expected end, got %v", child.Kind())
	}

}
