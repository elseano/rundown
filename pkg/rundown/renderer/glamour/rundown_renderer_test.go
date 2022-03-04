package glamour

import (
	"bytes"
	"testing"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/renderer"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Rundown struct {
	goldmark.Markdown

	Context *renderer.Context
}

func setupRenderer() *Rundown {
	context := renderer.NewContext("./virt")

	ansiOptions := ansi.Options{
		WordWrap:     80,
		ColorProfile: termenv.TrueColor,
	}

	rundownNodeRenderer := NewGlamourNodeRenderer(context)

	ar := ansi.NewRenderer(ansiOptions)
	renderer := goldrenderer.NewRenderer(
		goldrenderer.WithNodeRenderers(
			util.Prioritized(ar, 1000),
			util.Prioritized(rundownNodeRenderer, 1000),
		),
	)

	rundownRenderer := NewGlamourRenderer(
		renderer,
		context,
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

	return &Rundown{
		Markdown: gm,
		Context:  context,
	}
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

	gm := setupRenderer()

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

	gm := setupRenderer()
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

	gm := setupRenderer()
	doc := gm.Parser().Parse(text.NewReader(source))
	doc.Dump(source, 0)

	ast.PruneDocumentToSection(doc, "here")

	output := &bytes.Buffer{}

	if assert.NoError(t, gm.Renderer().Render(output, source, doc)) {
		assert.Equal(t, "\n\nStuff\n\n", output.String())
	}

}

func TestRenderContext(t *testing.T) {
	source1 := []byte(`

# Some Heading

<r capture-env/>

~~~ bash
export BLAH="Hi"
~~~

`)

	source2 := []byte(`
<r stdout nospin />

~~~ bash
sleep 1
echo $BLAH
~~~

`)

	gm := setupRenderer()
	context := gm.Context
	doc1 := gm.Parser().Parse(text.NewReader(source1))
	doc2 := gm.Parser().Parse(text.NewReader(source2))

	output := &bytes.Buffer{}

	if assert.NoError(t, gm.Renderer().Render(output, source1, doc1)) {
		assert.Equal(t, context.Env["BLAH"], "Hi")
		output := &bytes.Buffer{}

		if assert.NoError(t, gm.Renderer().Render(output, source2, doc2)) {
			assert.Equal(t, "  Hi\r\n\n", output.String())
		}
	}

}
