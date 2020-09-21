package markdown

import (
	"testing"

	"github.com/elseano/rundown/util"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

func TestCodeModifierBlock(t *testing.T) {
	contents1 := []byte("This is a test\n\n[](nospin)\n\n``` ruby\nSomecode\n```")
	verifyParsed(contents1, t)

	contents2 := []byte("This is a test\n\n[](nospin)\n``` ruby\nSomeboce\n```")
	verifyParsed(contents2, t)

	contents3 := []byte("This is a test\n\n<!--~ nospin -->\n``` ruby\nSomeboce\n```")
	verifyParsed(contents3, t)

	contents4 := []byte("This is a test\n\n<!---~ nospin --->\n``` ruby\nSomeboce\n```")
	verifyParsed(contents4, t)

}

func verifyParsed(contents []byte, t *testing.T) {
	markdown := goldmark.New(
		goldmark.WithExtensions(
			CodeModifiers,
		),
	)

	t.Log(string(contents))

	reader := text.NewReader([]byte(contents))
	doc := markdown.Parser().Parse(reader)

	dump := util.CaptureStdout(func() { doc.Dump(contents, 0) })
	t.Log(dump)

	child := doc.FirstChild()

	if child.Kind() != ast.KindParagraph {
		t.Fatalf("Expected KindParagraph, got %v", doc.FirstChild().Kind())
	}

	child = child.NextSibling()

	if child.Kind() != KindCodeModifierBlock {
		t.Fatal("Expected CodeModifierBlock")
	}

	child = child.NextSibling()

	if child.Kind() != ast.KindFencedCodeBlock {
		t.Fatal("Expected FencedCodeBlock")
	}

}
