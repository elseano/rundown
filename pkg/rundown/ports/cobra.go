package ports

import (
	"io/ioutil"
	"os"

	"github.com/charmbracelet/glamour/ansi"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/renderer"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

func BuildCobraCommand(sectionPointer *ast.SectionPointer) *cobra.Command {
	command := cobra.Command{
		Use:   sectionPointer.SectionName,
		Short: sectionPointer.DescriptionShort,
		Long:  sectionPointer.DescriptionLong,

		RunE: func(cmd *cobra.Command, args []string) error {
			ansiOptions := ansi.Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
			}

			source, err := ioutil.ReadFile(sectionPointer.OwningFilename)

			if err != nil {
				return err
			}

			rundownNodeRenderer := renderer.NewRundownNodeRenderer()

			ar := ansi.NewRenderer(ansiOptions)
			goldmarkRenderer := goldrenderer.NewRenderer(
				goldrenderer.WithNodeRenderers(
					util.Prioritized(ar, 1000),
					util.Prioritized(rundownNodeRenderer, 1000),
				),
			)

			rundownRenderer := renderer.NewRundownRenderer(goldmarkRenderer)

			gm := goldmark.New(
				goldmark.WithParserOptions(
					parser.WithASTTransformers(util.PrioritizedValue{
						Value:    transformer.NewRundownASTTransformer(),
						Priority: 0,
					}),
				),
				goldmark.WithRenderer(rundownRenderer),
			)

			doc := gm.Parser().Parse(text.NewReader(source))
			doc.Dump(source, 0)

			ast.PruneDocumentToSection(doc, sectionPointer.SectionName)

			return gm.Renderer().Render(os.Stdout, source, doc)
		},
	}

	return &command
}
