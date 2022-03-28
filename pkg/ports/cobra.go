package ports

import (
	"fmt"
	"os"
	"strings"

	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ast"

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

func BuildCobraCommand(filename string, section *rundown.Section, writeLog bool) *cobra.Command {
	optionEnv := map[string]optVal{}

	sectionPointer := section.Pointer
	doc := section.Document.Document
	source := section.Document.Source
	gm := section.Document.Goldmark

	longDesc := ""

	if sectionPointer.DescriptionLong != nil {
		str := strings.Builder{}
		gm.Renderer().Render(&str, source, sectionPointer.DescriptionLong)
		longDesc = str.String()
	}

	command := cobra.Command{
		Use:   sectionPointer.SectionName,
		Short: sectionPointer.DescriptionShort,
		Long:  longDesc,

		RunE: func(cmd *cobra.Command, args []string) error {
			if writeLog {
				devNull, _ := os.Create("rundown.log")
				rdutil.RedirectLogger(devNull)
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

			if val, err := cmd.Flags().GetBool("dump"); err == nil && val {
				doc.Dump(source, 1)
			}

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
