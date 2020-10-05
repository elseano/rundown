package markdown

import (
	"container/list"
	"strings"
	"testing"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"

	"github.com/elseano/rundown/pkg/util"

	"github.com/stretchr/testify/assert"
)

func TestRundownInline(t *testing.T) {
	contents := []byte("Normal <r>markdown</r> text")

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text", "..RundownInline", "...Text", "..Text"}, kindsFor(doc))
}

func TestRundownBlock(t *testing.T) {
	contents := []byte("<r some-attr some-val='val'>Rundown Block-like</r>")

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".RundownBlock", "..Paragraph", "...Text"}, kindsFor(doc))

	rundown := nodeAt(doc, 1).(*RundownBlock)
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

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text", ".ExecutionBlock"}, kindsFor(doc))

	eb := nodeAt(doc, 3).(*ExecutionBlock)
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

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text", ".FencedCodeBlock", ".ExecutionBlock"}, kindsFor(doc))
}

func TestExecutionBlockNorun(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash norun
echo Hello
~~~
	`)

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text"}, kindsFor(doc))
}

func TestExecutionBlockRevealNorun(t *testing.T) {
	contents := []byte(`
Paragraph

~~~ bash reveal norun
echo Hello
~~~
	`)

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text", ".FencedCodeBlock"}, kindsFor(doc))
}

func TestExecutionBlockRundown(t *testing.T) {
	contents := []byte(`
Paragraph

<r nospin/>

~~~ bash
echo Hello
~~~
	`)

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{"Document", ".Paragraph", "..Text", ".ExecutionBlock"}, kindsFor(doc))
	eb := nodeAt(doc, 3).(*ExecutionBlock)
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

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Heading",
		"....Text",
		"....RundownInline",
		"...Paragraph",
		"....Text",
		"...FencedCodeBlock",
		"...Section",
		"....Heading",
		".....Text",
		".....RundownInline",
		"....Paragraph",
		".....Text",
		"..Section",
		"...Heading",
		"....Text",
		"...Paragraph",
		"....Text",
	}, kindsFor(doc))

	assert.NotNil(t, nodeAt(doc, 2).(*Section).Label)
	assert.Equal(t, "one", *nodeAt(doc, 2).(*Section).Label)
	assert.NotNil(t, nodeAt(doc, 9).(*Section).Label)
	assert.Equal(t, "two", *nodeAt(doc, 9).(*Section).Label)
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

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Heading",
		"....Text",
		"....RundownInline",
		"...Paragraph",
		"....Text",
		"...RundownBlock",
		"....Paragraph",
		".....Text",
		"...FencedCodeBlock",
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

	markdown := PrepareMarkdown()
	reader := text.NewReader(contents)

	doc := markdown.Parser().Parse(reader)
	if doc.Parent() != nil {
		doc = doc.Parent()
	}

	t.Log(util.DumpNode(doc, contents))

	assert.Equal(t, []string{
		"SectionedDocument",
		".Document",
		"..Section",
		"...Heading",
		"....Text",
		"...ExecutionBlock",
	}, kindsFor(doc))

	assert.Equal(t, []string{"ExecutionBlock"}, kindsForList(nodeAt(doc, 2).(*Section).Setups))
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
