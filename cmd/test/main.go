package main

import (
	"io/ioutil"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	rr "github.com/elseano/rundown/pkg/rundown/renderer"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	rutil "github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

func main() {
	source, err := ioutil.ReadFile("./debug.md")
	if err != nil {
		panic(err)
	}

	file, err := os.Create("./debug.log")
	if err != nil {
		panic(err)
	}

	defer file.Close()
	rutil.RedirectLogger(file)

	// glamour.Render()

	ansiOptions := ansi.Options{
		WordWrap:     80,
		ColorProfile: termenv.TrueColor,
	}

	if termenv.HasDarkBackground() {
		ansiOptions.Styles = glamour.DarkStyleConfig
	} else {
		ansiOptions.Styles = glamour.LightStyleConfig
	}

	fileContext := rr.NewContext("./debug.md")
	rundownNodeRenderer := rr.NewRundownConsoleRenderer(fileContext)

	ar := ansi.NewRenderer(ansiOptions)
	renderer := renderer.NewRenderer(
		renderer.WithNodeRenderers(
			util.Prioritized(ar, 1000),
			util.Prioritized(rundownNodeRenderer, 1000),
		),
	)

	rundownRenderer := rr.NewRundownRenderer(
		renderer,
		fileContext,
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

	// doc := gm.Parser().Parse(text.NewReader(source))
	// doc.Dump(source, 0)

	gm.Convert(source, os.Stdout)

}
