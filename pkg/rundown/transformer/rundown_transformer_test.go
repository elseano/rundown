package transformer

import (
	"testing"

	"golang.org/x/net/html"

	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/stretchr/testify/assert"
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

	assert.Equal(t, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "subenv", Val: ""},
		},
		contents: "",
	}, extracted)
}

func TestHtmlExtractNested(t *testing.T) {
	rawHtml := goldast.NewRawHTML()
	rawHtml.Segments.Append(text.NewSegment(17, 21))
	source := text.NewReader([]byte("<r subenv>Context</r>"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Equal(t, &RundownHtmlTag{
		tag:      "r",
		attrs:    nil,
		contents: "",
		closed:   false,
		closer:   true,
	}, extracted)
}

func TestHtmlExtractBlock(t *testing.T) {
	rawHtml := goldast.NewHTMLBlock(goldast.HTMLBlockType1)
	rawHtml.Lines().Append(text.NewSegment(0, 21))
	source := text.NewReader([]byte("<r subenv>Context</r>"))

	extracted := ExtractRundownElement(rawHtml, source, "")

	assert.Equal(t, &RundownHtmlTag{
		tag: "r",
		attrs: []html.Attribute{
			{Namespace: "", Key: "subenv", Val: ""},
		},
		contents: "Context",
		closed:   true,
	}, extracted)
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

func TestExecutionBlockDefaults(t *testing.T) {
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

	if assert.NotNil(t, target) && assert.Equal(t, "ExecutionBlock", target.Kind().String()) {

		eb := target.(*ast.ExecutionBlock)

		assert.Equal(t, ast.SpinnerModeVisible, eb.SpinnerMode)
		assert.Equal(t, "Running...", eb.SpinnerName)

		assert.Equal(t, "go", eb.With)

		assert.Equal(t, false, eb.ShowStdout)

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

	if assert.NotNil(t, target) && assert.Equal(t, "SectionPointer", target.Kind().String()) {

		sp := target.(*ast.SectionPointer)

		assert.Equal(t, "SomeSection", sp.SectionName)
		assert.Equal(t, target.NextSibling(), sp.StartNode)
	}

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
			assert.Equal(t, "true", *sp.OptionDefault)
		}
	}

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

	if assert.NotNil(t, target) && assert.Equal(t, "DescriptionBlock", target.Kind().String()) {

		target = target.FirstChild()
		assert.Equal(t, "This is some description", string(target.Text(source)))

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
