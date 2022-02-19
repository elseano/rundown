package ports

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/renderer"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type optVal struct {
	Str    *string
	Bool   *bool
	Option *ast.SectionOption
}

func (o optVal) String() string {
	if o.Str != nil {
		return *o.Str
	}

	if o.Bool != nil {
		if *o.Bool {
			return "true"
		} else {
			return "false"
		}
	}

	return ""
}

func BuildCobraCommand(filename string, sectionPointer *ast.SectionPointer, debugAs string) *cobra.Command {
	optionEnv := map[string]optVal{}

	command := cobra.Command{
		Use:   sectionPointer.SectionName,
		Short: sectionPointer.DescriptionShort,
		Long:  sectionPointer.DescriptionLong,

		RunE: func(cmd *cobra.Command, args []string) error {
			if debugAs == "" {
				devNull, _ := os.Create("rundown.log")
				rdutil.RedirectLogger(devNull)
			} else {
				rdutil.RedirectLogger(os.Stdout)
				rdutil.SetLoggerLevel(debugAs)
			}

			ansiOptions := ansi.Options{
				WordWrap:     80,
				ColorProfile: termenv.TrueColor,
				Styles:       glamour.DarkStyleConfig,
			}

			source, err := ioutil.ReadFile(filename)

			if err != nil {
				return err
			}

			executionContext := renderer.NewContext(filename)
			executionContext.ImportRawEnv(os.Environ())

			rundownNodeRenderer := renderer.NewRundownConsoleRenderer(executionContext)

			ar := ansi.NewRenderer(ansiOptions)
			goldmarkRenderer := goldrenderer.NewRenderer(
				goldrenderer.WithNodeRenderers(
					util.Prioritized(ar, 1000),
					util.Prioritized(rundownNodeRenderer, 1),
				),
			)

			for k, v := range optionEnv {
				if err := v.Option.OptionType.Validate(v.String()); err != nil {
					return fmt.Errorf("%s: %w", v.Option.OptionName, err)
				}

				executionContext.ImportEnv(map[string]string{
					k: fmt.Sprintf("%v", v.String()),
				})
			}

			rundownRenderer := renderer.NewRundownRenderer(goldmarkRenderer, executionContext)

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

			ast.PruneDocumentToSection(doc, sectionPointer.SectionName)
			doc.Dump(source, 0)

			return gm.Renderer().Render(os.Stdout, source, doc)
			// return nil
		},
	}

	for _, o := range sectionPointer.Options {
		opt := o
		switch topt := opt.OptionType.(type) {
		case *ast.TypeString:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
		case *ast.TypeBoolean:
			optionEnv[opt.OptionAs] = optVal{Bool: command.Flags().Bool(opt.OptionName, topt.Normalise(opt.OptionDefault.String) == "true", opt.OptionDescription), Option: opt}
		case *ast.TypeEnum:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
		case *ast.TypeFilename:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
		}

		if opt.OptionRequired {
			command.MarkFlagRequired(opt.OptionName)
		}

	}

	command.SetFlagErrorFunc(func(errCmd *cobra.Command, err error) error {
		return fmt.Errorf("ERROR")
	})

	return &command
}
