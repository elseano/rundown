package ports

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"path"

	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/errs"
	"gopkg.in/guregu/null.v4"

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
		Use:           sectionPointer.SectionName,
		Short:         sectionPointer.DescriptionShort,
		Long:          longDesc,
		SilenceErrors: true,

		RunE: func(cmd *cobra.Command, args []string) error {
			if writeLog {
				devNull, _ := os.Create("rundown.log")
				rdutil.RedirectLogger(devNull)
			}

			executionContext := section.Document.Context
			executionContext.ImportRawEnv(os.Environ())
			executionContext.RundownFile = section.Document.Filename

			for k, v := range optionEnv {
				if err := v.Option.OptionType.Validate(v.String()); err != nil {
					return fmt.Errorf("%s: %w", v.Option.OptionName, err)
				}

				executionContext.ImportEnv(map[string]string{
					k: fmt.Sprintf("%v", v.String()),
				})
			}

			ast.PruneDocumentToSection(doc, sectionPointer.SectionName)
			sectionPointer.SetIfScript("") // Ensure the requested section runs.

			if val, err := cmd.Flags().GetBool("dump"); err == nil && val {
				doc.Dump(source, 1)
			}

			fmt.Printf("Running %s in %s...\n\n", sectionPointer.SectionName, section.Document.Filename)

			executionContext.ImportEnv(map[string]string{"PWD": path.Dir(executionContext.RundownFile)})


			err := gm.Renderer().Render(os.Stdout, source, doc)

			switch {
			case errors.Is(err, errs.ErrStopOk):
				return nil
			default:
				return err
			}
		},
	}

	for _, o := range sectionPointer.Options {
		opt := o

		// Populate defaults with the current environment setting, unless default has been specified.
		if cval, ok := os.LookupEnv(opt.OptionAs); ok && !opt.OptionDefault.Valid {
			opt.OptionDefault = null.NewString(cval, cval != "")
		}

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

		if opt.OptionRequired && !opt.OptionDefault.Valid {
			command.MarkFlagRequired(opt.OptionName)
		}

	}

	command.SetFlagErrorFunc(func(errCmd *cobra.Command, err error) error {
		fmt.Printf("Error: %s\n", err.Error())
		return fmt.Errorf("error: %w", err)
	})

	return &command
}
