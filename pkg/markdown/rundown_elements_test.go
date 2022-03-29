package markdown

import (
	"container/list"
	"strings"
	"testing"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"

	"github.com/elseano/rundown/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestRundownInline(t *testing.T) {
	contents := []byte("Normal <r>markdown</r> text")

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text", "....RundownInline", ".....Text", "....Text"}, kindsFor(doc))

	section, ok := nodeAt(doc, 1).(*Section)
	assert.True(t, ok)
	assert.Equal(t, "Root", section.Name)
}

func TestRundownBlock(t *testing.T) {
	contents := []byte("<r some-attr some-val='val'>Rundown Block-like</r>")

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...RundownBlock", "....Paragraph", ".....Text"}, kindsFor(doc))

	rundown := nodeAt(doc, 3).(*RundownBlock)
	assert.True(t, rundown.Modifiers.Flags[Flag("some-attr")])
	assert.Equal(t, rundown.Modifiers.Values[Parameter("some-val")], "val")
}

func TestExecutionBlock(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash nospin
echo Hello
~~~
	`)

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text", "..ExecutionBlock"}, kindsFor(doc))

	eb := nodeAt(doc, 5).(*ExecutionBlock)
	assert.Equal(t, "bash", eb.Syntax)
	assert.True(t, eb.Modifiers.Flags[Flag("nospin")])
}

func TestExecutionBlockReveal(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash reveal
echo Hello
~~~
	`)

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text", "...FencedCodeBlock", "..ExecutionBlock"}, kindsFor(doc))
}

func TestExecutionBlockNorun(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash norun
echo Hello
~~~
	`)

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text"}, kindsFor(doc))
}

func TestExecutionBlockRevealNorun(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash reveal norun
echo Hello
~~~
	`)

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text", "...FencedCodeBlock"}, kindsFor(doc))
}

func TestExecutionBlockRundown(t *testing.T) {
	contents := []byte(`
Paragraph

<r nospin/>

~~~ bash
echo Hello
~~~
	`)

	doc := getAst(contents)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Section", "..Document", "...Paragraph", "....Text", "..ExecutionBlock"}, kindsFor(doc))
	eb := nodeAt(doc, 5).(*ExecutionBlock)
	assert.True(t, eb.Modifiers.Flags[Flag("nospin")])
}

func TestSection(t *testing.T) {
	contents := []byte(`
# Heading <r label="one"/>

Contents within section

~~~
echo Blah
~~~

## SubSection <r label="two"/>

Child Section

# Back to Normal

Blah

	`)

	doc := getAst(contents)

	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Document",
		"....Heading",
		".....Text",
		".....RundownInline",
		"....Paragraph",
		".....Text",
		"....FencedCodeBlock",
		"...Section",
		"....Document",
		".....Heading",
		"......Text",
		"......RundownInline",
		".....Paragraph",
		"......Text",
		"..Section",
		"...Document",
		"....Heading",
		".....Text",
		"....Paragraph",
		".....Text",
	}, kindsFor(doc))

	assert.NotEqual(t, "Root", nodeAt(doc, 2).(*Section).Name)
	assert.NotNil(t, nodeAt(doc, 2).(*Section).Label)
	assert.Equal(t, "one", *nodeAt(doc, 2).(*Section).Label)
	assert.NotNil(t, nodeAt(doc, 10).(*Section).Label)
	assert.Equal(t, "two", *nodeAt(doc, 10).(*Section).Label)
}

func TestSectionDesc(t *testing.T) {
	contents := []byte(`
# Heading <r label="one"/>

Contents within section

<r desc>This is a description of the shortcode</r>

~~~
echo Blah
~~~
	`)

	doc := getAst(contents)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Document",
		"....Heading",
		".....Text",
		".....RundownInline",
		"....Paragraph",
		".....Text",
		"....RundownBlock",
		".....Paragraph",
		"......Text",
		"....FencedCodeBlock",
	}, kindsFor(doc))

	desc := nodeAt(doc, 2).(*Section).Description
	assert.Equal(t, 1, desc.Len())
	assert.Equal(t, "This is a description of the shortcode", string(desc.Front().Value.(*RundownBlock).Text(contents)))

}

func TestSectionSetup(t *testing.T) {
	contents := []byte(`
# Heading

~~~ bash setup
echo Hi
~~~

	`)

	doc := getAst(contents)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Document",
		"....Heading",
		".....Text",
		"...ExecutionBlock",
	}, kindsFor(doc))

	assert.Equal(t, []string{"ExecutionBlock"}, kindsForList(nodeAt(doc, 2).(*Section).Setups))
}

func TestSectionHeadingSkips(t *testing.T) {
	contents := []byte(`
# Heading

One.

#### Heading 2

Two.

#### Heading 3

Three.

## Heading 4

Four.
	`)

	doc := getAst(contents)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Document",
		"....Heading",
		".....Text",
		"....Paragraph",
		".....Text",
		"...Section",
		"....Document",
		".....Heading",
		"......Text",
		".....Paragraph",
		"......Text",
		"...Section",
		"....Document",
		".....Heading",
		"......Text",
		".....Paragraph",
		"......Text",
		"...Section",
		"....Document",
		".....Heading",
		"......Text",
		".....Paragraph",
		"......Text",
	}, kindsFor(doc))

	assert.Equal(t, 1, nodeAt(doc, 2).(*Section).Level)
	assert.Equal(t, 4, nodeAt(doc, 8).(*Section).Level)
	assert.Equal(t, 4, nodeAt(doc, 14).(*Section).Level)
	assert.Equal(t, 2, nodeAt(doc, 20).(*Section).Level)
}

func nodeAt(node ast.Node, index int) ast.Node {
	var counter = 0
	var candidate ast.Node = nil

	ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			counter++
			if counter > index {
				candidate = child
				return ast.WalkStop, nil
			}
		}

		return ast.WalkContinue, nil
	})

	return candidate
}

func kindsFor(node ast.Node) []string {
	var result = []string{}
	var depth = 0

	ast.Walk(node, func(child ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			result = append(result, strings.Repeat(".", depth)+child.Kind().String())
			depth++
		} else {
			depth--
		}

		return ast.WalkContinue, nil
	})

	return result
}

func kindsForList(nodes list.List) []string {
	var result = []string{}

	for node := nodes.Front(); node != nil; node = node.Next() {
		for _, k := range kindsFor(node.Value.(ast.Node)) {
			result = append(result, k)
		}
	}

	return result
}

func getAst(contents []byte) ast.Node {

	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Strikethrough,
			RundownElements,
			Emoji,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
	)

	reader := text.NewReader(contents)

	doc := md.Parser().Parse(reader)

	return doc
}
