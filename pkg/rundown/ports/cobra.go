package ports

import (
	"fmt"
	"os"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/rundown/ast"

	// glamrend "github.com/elseano/rundown/pkg/rundown/renderer/glamour"

	rdutil "github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
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

func BuildCobraCommand(filename string, section *rundown.Section, debugAs string) *cobra.Command {
	optionEnv := map[string]optVal{}

	sectionPointer := section.Pointer
	doc := section.Document.Document
	source := section.Document.Source
	gm := section.Document.Goldmark

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

			executionContext := section.Document.Context
			executionContext.ImportRawEnv(os.Environ())

			for k, v := range optionEnv {
				if err := v.Option.OptionType.Validate(v.String()); err != nil {
					return fmt.Errorf("%s: %w", v.Option.OptionName, err)
				}

				executionContext.ImportEnv(map[string]string{
					k: fmt.Sprintf("%v", v.String()),
				})
			}

			ast.PruneDocumentToSection(doc, sectionPointer.SectionName)
			doc.Dump(source, 1)

			ast := rdutil.CaptureStdout(func() {
				doc.Dump(source, 0)
			})

			rdutil.Logger.Debug().Msg(ast)

			return gm.Renderer().Render(os.Stdout, source, doc)
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
