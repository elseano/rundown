package transformer

import (
	"fmt"
	"strings"
	"testing"

	"golang.org/x/net/html"

	"github.com/elseano/rundown/pkg/ast"
	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yuin/goldmark"
	goldast "github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
	// "bytes"
	// "io/ioutil"
)

func TestSubEnv(t *testing.T) {
	reader := text.NewReader([]byte("Text <r subenv>Content $SUBENV and more **$CONTENT**</r> More $TEXT"))
	doc := goldmark.DefaultParser().Parse(reader)

	result := NewRundownASTTransformer()
	result.Transform(doc.(*goldast.Document), reader, parser.NewContext())

	doc.Dump(reader.Source(), 0)

	target := doc.FirstChild().FirstChild().NextSibling()

	assert.Equal(t, "Text", target.Kind().String())
	assert.Equal(t, "Content ", string(target.Text(reader.Source())))

	target = target.NextSibling()

	assert.Equal(t, "EnvironmentSubstitution", target.Kind().String())
	assert.Equal(t, "$SUBENV", string(target.Text(reader.Source())))

	target = target.NextSibling()

	assert.Equal(t, "Text", target.Kind().String())
	assert.Equal(t, " and more ", string(target.Text(reader.Source())))

	target = target.NextSibling()

	assert.Equal(t, "Emphasis", target.Kind().String())
	// assert.Equal(t, "$CONTENT", string(target.Text(reader.Source())))

	target = target.FirstChild()

	assert.Equal(t, "EnvironmentSubstitution", target.Kind().String())
	assert.Equal(t, "$CONTENT", string(target.Text(reader.Source())))

	target = target.Parent().NextSibling()

	assert.Equal(t, "Text", target.Kind().String())
	assert.Equal(t, " More $TEXT", string(target.Text(reader.Source())))

}

func TestHtmlExtractInlineOpening(t *testing.T) {
	rawHtml := goldast.NewRawHTML()
	rawHtml.Segments.Append(text.NewSegment(0, 10))
	source := text.NewReader([]byte("<r subenv>Context</r>"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Contains(t, extracted, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "subenv", Val: ""},
		},
		contents: "",
	})
}

func TestHtmlExtractNested(t *testing.T) {
	rawHtml := goldast.NewRawHTML()
	rawHtml.Segments.Append(text.NewSegment(17, 21))
	source := text.NewReader([]byte("<r subenv>Context</r>"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Contains(t, extracted, &RundownHtmlTag{
		tag:      "r",
		attrs:    nil,
		contents: "",
		closed:   false,
		closer:   true,
	})
}

func TestHtmlExtractBlock(t *testing.T) {
	rawHtml := goldast.NewHTMLBlock(goldast.HTMLBlockType1)
	rawHtml.Lines().Append(text.NewSegment(0, 21))
	source := text.NewReader([]byte("<r subenv>Context</r>"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Contains(t, extracted, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "subenv", Val: ""},
		},
		contents: "Context",
		closed:   true,
	})
}

func TestHtmlExtractSequential(t *testing.T) {
	rawHtml := goldast.NewHTMLBlock(goldast.HTMLBlockType1)
	rawHtml.Lines().Append(text.NewSegment(0, 35))
	source := text.NewReader([]byte("<r something /><r something-else />"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Contains(t, extracted, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "something", Val: ""},
		},
		closed: true,
	})

	assert.Contains(t, extracted, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "something-else", Val: ""},
		},
		closed: true,
	})
}

func TestRundownBlockFlattened(t *testing.T) {
	source := []byte("<r import='blah'>[Something](./something.md)</r>")

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	assertTree(t, doc, []string{"Document", " ImportBlock"})
}

func TestRundownBlockComplex(t *testing.T) {
	source := []byte("<r help>\n\nHere's `something`\n\nAnd another thing\n</r>\n\nAnd then some stuff.")

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))
	doc.Dump(source, 0)

	assertTree(t, doc, []string{"Document", " DescriptionBlock", "  Paragraph", "  Paragraph", " Paragraph"})
}

func TestExecutionBlockSpecified(t *testing.T) {
	source := []byte(`
<r spinner="Some spinner name..." with="go run $FILE" stdout />

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	assert.Equal(t, "ExecutionBlock", target.Kind().String())

	eb := target.(*ast.ExecutionBlock)

	assert.Equal(t, ast.SpinnerModeVisible, eb.SpinnerMode)
	assert.Equal(t, "Some spinner name...", eb.SpinnerName)

	assert.Equal(t, "go run $FILE", eb.With)
	assert.Equal(t, true, eb.ShowStdout)

	target = target.NextSibling()

	// Ensure fenced code block is removed from tree.
	assert.Nil(t, target)

}

func TestExecutionBlockNotDefault(t *testing.T) {
	source := []byte(`
~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	assert.NotNil(t, target)
	assert.Equal(t, "FencedCodeBlock", target.Kind().String())

}

func TestExecutionBlockReveal(t *testing.T) {
	source := []byte(`
<r reveal/>

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "FencedCodeBlock", target.Kind().String()) {

		target = target.NextSibling()

		if assert.NotNil(t, target) && assert.Equal(t, "ExecutionBlock", target.Kind().String()) {
			eb := target.(*ast.ExecutionBlock)

			assert.Equal(t, ast.SpinnerModeVisible, eb.SpinnerMode)
			assert.Equal(t, "Running...", eb.SpinnerName)

			assert.Equal(t, "go", eb.With)
			assert.Equal(t, true, eb.Reveal)

			assert.Equal(t, false, eb.ShowStdout)

		}
	}

}

func TestSaveCodeBlock(t *testing.T) {
	source := []byte(`
<r save="GOSRC"/>

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "SaveCodeBlock", target.Kind().String()) {

		scb := target.(*ast.SaveCodeBlock)

		assert.Equal(t, "GOSRC", scb.SaveToVariable)

	}

}

func TestSectionInsideHeading(t *testing.T) {
	source := []byte(`
# This is a heading <r section="SomeSection"/>

<r desc="This is a longer description"/>

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	fmt.Printf("RESULT:\n")
	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "SectionPointer", target.Kind().String()) {

		sp := target.(*ast.SectionPointer)

		assert.Equal(t, "SomeSection", sp.SectionName)
		require.Equal(t, 3, target.ChildCount())
		assert.Equal(t, target.FirstChild().Kind(), goldast.KindHeading)
		assert.Equal(t, "This is a heading", sp.DescriptionShort)
		// assert.Equal(t, "This is a longer description", string(sp.DescriptionLong.FirstChild().(*goldast.String).Value))
	}

}

func TestSectionTermination(t *testing.T) {
	source := []byte(`
# This is a heading <r section="SomeSection"/>

I'm SomeSection.

## SubSection <r section="SomeSection:Sub"/>

I'm a sub section inside SomSection.

# Another heading <r section="SomeOtherSection"/>

I'm SomeOtherSection.

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	t.Logf("RESULT:\n")

	t.Logf("AST: %s", rdutil.CaptureStdout(func() {
		doc.Dump(source, 0)
	}))

	sections := ast.GetSections(doc)

	require.Equal(t, 3, len(sections))

	require.Equal(t, 3, sections[0].ChildCount())
	require.Equal(t, 2, sections[1].ChildCount())
	require.Equal(t, 2, sections[2].ChildCount())
}

func TestSectionOption(t *testing.T) {
	source := []byte(`

<r opt="skip-production" type="bool" default="true"/>

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "SectionOption", target.Kind().String()) {

		sp := target.(*ast.SectionOption)

		assert.Equal(t, "skip-production", sp.OptionName)
		assert.Equal(t, &ast.TypeBoolean{}, sp.OptionType)

		if assert.NotNil(t, sp.OptionDefault) {
			assert.Equal(t, "true", sp.OptionDefault.String)
		}
	}

}

func TestSectionOptionInsideSection(t *testing.T) {
	source := []byte(`
## Blah <r section="test">

<r opt="skip-production" type="bool" default="true"/>

</r>

~~~ go
blah
~~~
`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := ast.FindSectionInDocument(doc, "test")

	if assert.NotNil(t, target) {
		require.Len(t, target.Options, 1)
		assert.Equal(t, "skip-production", target.Options[0].OptionName)
	}

}

func TestSectionDependencies(t *testing.T) {
	source := []byte(`

## Some dep <r section="dep1" />

<r spinner="Something"/>

~~~ bash
echo "Hi"
~~~

<r invoke="dep2" title="Blah" />

## Blah <r section="test"/>

<r dep="dep1"/>

<r spinner="Blah..."/>

~~~ go
blah
~~~

# Dep2 <r section="dep2"/>

<r opt="title" type="string" required desc="title"/>

Some secondary dep.

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	ast.PruneDocumentToSection(doc, "test")
	target := ast.FindSectionInDocument(doc, "test")

	doc.Dump(source, 0)

	if assert.NotNil(t, target) {
		require.Len(t, target.Dependencies, 2)
		assert.Equal(t, "dep1", target.Dependencies[0].SectionName)
		assert.Equal(t, "dep2", target.Dependencies[1].SectionName)

		require.Len(t, target.Dependencies[1].Options, 1)
		assert.Equal(t, "title", target.Dependencies[1].Options[0].OptionName)

		require.Equal(t, 3, target.Dependencies[0].ChildCount())
		assert.Equal(t, "Heading", target.Dependencies[0].FirstChild().Kind().String())
		assert.Equal(t, "ExecutionBlock", target.Dependencies[0].FirstChild().NextSibling().Kind().String())
		assert.Equal(t, "InvokeBlock", target.Dependencies[0].FirstChild().NextSibling().NextSibling().Kind().String())

		assertTree(t, target, []string{
			"SectionPointer",
			" Heading",
			" InvokeBlock",
			"  Heading",
			"  ExecutionBlock",
			"  InvokeBlock",
			"   Heading",
			"   Paragraph",
			" ExecutionBlock",
		})
	}
}

func writeChildren(b *strings.Builder, n goldast.Node, depth int) {
	type inline interface{ Inline() }

	for child := n.FirstChild(); child != nil; child = child.NextSibling() {
		if _, isInline := child.(inline); !isInline {
			b.WriteString(strings.Repeat(" ", depth))
			b.WriteString(fmt.Sprintf("%s\n", child.Kind().String()))

			writeChildren(b, child, depth+1)
		}
	}
}

func assertTree(t *testing.T, node goldast.Node, tree []string) {
	b := strings.Builder{}

	b.WriteString(fmt.Sprintf("%s\n", node.Kind().String()))

	writeChildren(&b, node, 1)

	assert.Equal(t, strings.Join(tree, "\n"), strings.TrimSpace(b.String()))
}

func TestDescriptionAttr(t *testing.T) {
	source := []byte(`

<r desc="Some description"/>

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "DescriptionBlock", target.Kind().String()) {

	}

}

func TestDescriptionBlock(t *testing.T) {
	source := []byte(`

<r desc>This is some description</r>

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "Paragraph", target.Kind().String()) {

		target := target.FirstChild()

		if assert.NotNil(t, target) && assert.Equal(t, "DescriptionBlock", target.Kind().String()) {

			target = target.FirstChild()
			assert.Equal(t, "This is some description", string(target.Text(source)))

		}

	}

}

func TestStopFailAttr(t *testing.T) {
	source := []byte(`

<r stop-fail="Some reason"/>

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "StopFail", target.Kind().String()) {

		target = target.FirstChild()
		assert.Equal(t, "Some reason", string(target.Text(source)))

	}

}

func TestStopFailBlock(t *testing.T) {
	source := []byte(`

<r stop-fail>Some Reason</r>

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	if assert.NotNil(t, target) && assert.Equal(t, "StopFail", target.Kind().String()) {

		target = target.FirstChild()
		assert.Equal(t, "Some Reason", string(target.Text(source)))

	}

}

func TestIgnoreBlock(t *testing.T) {
	source := []byte(`

<r ignore>This is some text to ignore.</r>

`)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
	)

	doc := gm.Parser().Parse(text.NewReader(source))

	doc.Dump(source, 0)

	target := doc.FirstChild()

	assert.Nil(t, target)

}
