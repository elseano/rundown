package ports

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/errs"
	"github.com/elseano/rundown/pkg/renderer/term"
	"github.com/muesli/reflow/indent"
	"golang.org/x/exp/maps"
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

	return "<err>"
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
		writer := indent.NewWriterPipe(&str, 2, nil)
		gm.Renderer().Render(writer, source, sectionPointer.DescriptionLong)
		longDesc = sectionPointer.DescriptionShort + "\n\n" + str.String()
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

			optionEnvStr := map[string]string{}
			for k, v := range optionEnv {
				fmt.Printf("Setting %s to %s\n", k, v.String())
				optionEnvStr[k] = v.String()
			}

			parsed, err := sectionPointer.ParseOptions(optionEnvStr)

			if err != nil {
				return err
			}

			executionContext.ImportEnv(parsed)

			if err := ast.FillInvokeBlocks(doc, 10); err != nil {
				return err
			}

			doc = ast.PruneDocumentToSection(doc, sectionPointer.SectionName)
			sectionPointer.SetIfScript("") // Ensure the requested section runs.

			if val, err := cmd.Flags().GetBool("dump"); err == nil && val {
				doc.Dump(source, 1)
			}

			out := rdutil.CaptureStdout(func() {
				doc.Dump(source, 0)
			})

			rdutil.Logger.Debug().Msg(out)

			rdutil.Logger.Info().Msgf("Running %s in %s...\n\n", sectionPointer.SectionName, section.Document.Filename)

			executionContext.ImportEnv(map[string]string{"PWD": path.Dir(executionContext.RundownFile)})

			err = gm.Renderer().Render(os.Stdout, source, doc)

			switch {
			case errors.Is(err, errs.ErrStopOk):
				return nil
			default:
				return err
			}
		},
	}

	var flagsRequired []string

	for _, o := range sectionPointer.Options {
		opt := o

		// Populate defaults with the current environment setting, unless default has been specified.
		if cval, ok := os.LookupEnv(opt.OptionAs); ok && !opt.OptionDefault.Valid {
			opt.OptionDefault = null.NewString(cval, cval != "")
		}

		switch topt := opt.OptionType.(type) {
		case *ast.TypeString:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
			command.RegisterFlagCompletionFunc(opt.OptionName, stringCompletionFunction(opt, topt))
			command.Flags()
		case *ast.TypeBoolean:
			optionEnv[opt.OptionAs] = optVal{Bool: command.Flags().Bool(opt.OptionName, topt.Normalise(opt.OptionDefault.String) == "true", opt.OptionDescription), Option: opt}
			command.RegisterFlagCompletionFunc(opt.OptionName, boolCompletionFunction(opt, topt))
		case *ast.TypeEnum:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
			command.RegisterFlagCompletionFunc(opt.OptionName, enumCompletionFunction(topt))
		case *ast.TypeFilename:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
			command.RegisterFlagCompletionFunc(opt.OptionName, filenameCompletionFunction(topt))
		case *ast.TypeKV:
			optionEnv[opt.OptionAs] = optVal{Str: command.Flags().String(opt.OptionName, opt.OptionDefault.String, opt.OptionDescription), Option: opt}
			command.RegisterFlagCompletionFunc(opt.OptionName, kvCompletionFunction(topt))

		}

		if opt.OptionRequired && !opt.OptionDefault.Valid {
			command.MarkFlagRequired(opt.OptionName)

			flagsRequired = append(flagsRequired, term.Aurora.BrightYellow("--"+opt.OptionName).String())
			command.Use = command.Use + " --" + opt.OptionName + " (" + opt.OptionType.InputType() + ")"
		}

	}

	command.SetFlagErrorFunc(func(errCmd *cobra.Command, err error) error {
		fmt.Println(command.Help())
		fmt.Println()
		// display := term.Aurora.Sprintf("Running %s requires flags %s", term.Aurora.BrightCyan(sectionPointer.SectionName), strings.Join(flagsRequired, ", "))
		return err
	})

	return &command
}

func enumCompletionFunction(opt *ast.TypeEnum) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return opt.ValidValues, cobra.ShellCompDirectiveNoFileComp
	}
}

func kvCompletionFunction(opt *ast.TypeKV) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return maps.Keys(opt.Pairs), cobra.ShellCompDirectiveNoFileComp
	}
}

func stringCompletionFunction(opt *ast.SectionOption, topt *ast.TypeString) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// return cobra.AppendActiveHelp([]string{"xx"}, opt.OptionDescription), cobra.ShellCompDirectiveNoFileComp
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
}

func boolCompletionFunction(opt *ast.SectionOption, topt *ast.TypeBoolean) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// return cobra.AppendActiveHelp([]string{"true", "false"}, opt.OptionDescription), cobra.ShellCompDirectiveNoFileComp
		return []string{"true", "false"}, cobra.ShellCompDirectiveNoFileComp
	}
}

func filenameCompletionFunction(opt *ast.TypeFilename) func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if opt.MustExist {
		return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// return cobra.AppendActiveHelp([]string{""}, "Requires a file which exists"), cobra.ShellCompDirectiveDefault
			return nil, cobra.ShellCompDirectiveDefault
		}
	} else {
		return func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			// return cobra.AppendActiveHelp([]string{""}, "Requires a file which does not exist"), cobra.ShellCompDirectiveNoFileComp
			return nil, cobra.ShellCompDirectiveNoFileComp
		}
	}
}
