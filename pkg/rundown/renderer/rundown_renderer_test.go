package renderer

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func setupRenderer(source []byte) goldmark.Markdown {
	ansiOptions := ansi.Options{
		WordWrap:     80,
		ColorProfile: termenv.TrueColor,
	}

	rundownNodeRenderer := NewRundownNodeRenderer()

	ar := ansi.NewRenderer(ansiOptions)
	renderer := renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(ar, 1000),
			util.Prioritized(rundownNodeRenderer, 1000),
		),
	)

	rundownRenderer := NewRundownRenderer(
		renderer,
	)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(util.PrioritizedValue{
				Value:    transformer.NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
		goldmark.WithRenderer(rundownRenderer),
	)

	return gm
}

func TestNormalRender(t *testing.T) {
	source := []byte(`

# Some Heading

Blah

<r reveal/>

~~~ bash
echo "Hi"
~~~

Done.

`)

	gm := setupRenderer(source)

	doc := gm.Parser().Parse(text.NewReader(source))
	doc.Dump(source, 0)

	output := &bytes.Buffer{}

	if assert.NoError(t, gm.Convert(source, output)) {
		assert.Equal(t, "Some Heading\n\nBlah\n\n\necho \"Hi\"\n\n\n\nDone.\n\n", output.String())
	}
}

func TestJumpRenderHeading(t *testing.T) {
	source := []byte(`

# Some Heading

Blah

<r reveal/>

~~~ bash
echo "Hi"
~~~

Done.

# Some other heading <r section="here"/>

Stuff

## Subheading

# Yet another heading

More stuff.

`)

	gm := setupRenderer(source)
	doc := gm.Parser().Parse(text.NewReader(source))
	doc.Dump(source, 0)

	ast.PruneDocumentToSection(doc, "here")

	output := &bytes.Buffer{}

	if assert.NoError(t, gm.Renderer().Render(output, source, doc)) {
		assert.Equal(t, "\n\nSome other heading\n\nStuff\n\n\nSubheading\n", output.String())
	}

}

func TestJumpRenderBlock(t *testing.T) {
	source := []byte(`

# Some Heading

Blah

<r reveal/>

~~~ bash
echo "Hi"
~~~

Done.

# Some other heading

<r section="here" title="Do a thing">

Stuff

</r>

## Subheading

# Yet another heading

More stuff.

`)

	gm := setupRenderer(source)
	doc := gm.Parser().Parse(text.NewReader(source))
	doc.Dump(source, 0)

	ast.PruneDocumentToSection(doc, "here")

	output := &bytes.Buffer{}

	if assert.NoError(t, gm.Renderer().Render(output, source, doc)) {
		assert.Equal(t, "\n\nStuff\n\n", output.String())
	}

}
