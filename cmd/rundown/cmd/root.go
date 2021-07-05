package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	shared "github.com/elseano/rundown/cmd"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/rundown/ast"
	"github.com/elseano/rundown/pkg/rundown/ports"
	"github.com/elseano/rundown/pkg/rundown/renderer"
	"github.com/elseano/rundown/pkg/rundown/transformer"
	"github.com/elseano/rundown/pkg/util"
	"github.com/muesli/termenv"
	"github.com/spf13/cobra"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/parser"
	goldrenderer "github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	goldutil "github.com/yuin/goldmark/util"
)

func Execute(version string, gitCommit string) error {
	cmd := NewDocRootCmd(os.Args)
	cmd.Version = version + " (" + gitCommit + ")"

	return cmd.Execute()
}

var positionalArgs = []*rundown.ShortCodeOption{}

func validatePositional(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return []string{}, cobra.ShellCompDirectiveDefault
}

func addArgs(cmd *cobra.Command, options map[string]*rundown.ShortCodeOption) {
	argsCount := 0

	argLines := ""
	maxOptLength := 0

	for _, opt := range options {
		if opt.IsPositional {
			if len(opt.Code) > maxOptLength {
				maxOptLength = len(opt.Code)
			}
		}
	}

	fmtString := fmt.Sprintf("%%s\n  %%-%ds   %%s", maxOptLength)

	for _, opt := range options {
		if opt.IsPositional {
			argsCount++

			positionalArgs = append(positionalArgs, opt)

			argLines = fmt.Sprintf(fmtString, argLines, opt.Code, opt.Description)

			if opt.Position == -1 {
				cmd.Use = fmt.Sprintf("%s <%s>...", cmd.Use, opt.Code)
			} else {
				cmd.Use = fmt.Sprintf("%s <%s>", cmd.Use, opt.Code)
			}
		}
	}

	cmd.ValidArgsFunction = validatePositional
	cmd.Use = fmt.Sprintf("%s\n%s", cmd.Use, argLines)

	if argsCount == -1 {
		cmd.Args = cobra.ArbitraryArgs
	} else {
		cmd.Args = cobra.OnlyValidArgs
	}
}

func addFlags(cmd *cobra.Command, options map[string]*rundown.ShortCodeOption) {
	for _, opt := range options {
		var description = opt.Description
		var isRequired = opt.Required && opt.Default == ""
		var willAsk = !flagNonInteractive && opt.Prompt

		if isRequired {
			description = fmt.Sprintf("%s [REQUIRED]", description)
		}

		if !opt.IsPositional {
			switch opt.Type {
			case "string":
				cmd.Flags().String(opt.Code, opt.Default, description)
			case "file-exists":
				cmd.Flags().String(opt.Code, opt.Default, description)
				cmd.MarkFlagFilename(opt.Code)
			case "file-not-exists":
				cmd.Flags().String(opt.Code, opt.Default, description)
			case "bool":
				cmd.Flags().Bool(opt.Code, false, description)
			case "int":
				cmd.Flags().Int32(opt.Code, -1, description)
			default:
				cmd.Flags().String(opt.Code, opt.Default, description)
			}

			if isRequired && !willAsk {
				cmd.MarkFlagRequired(opt.Code)
			}
		}
	}
}

func NewDocRootCmd(args []string) *cobra.Command {
	docRoot := NewRootCmd()
	docRoot.ParseFlags(args)

	rundownFile = shared.RundownFile(flagFilename)
	// cwd, _ := os.Getwd()
	// relRundownFile, _ := filepath.Rel(cwd, rundownFile)

	ansiOptions := ansi.Options{
		WordWrap:     80,
		ColorProfile: termenv.TrueColor,
		Styles:       glamour.DarkStyleConfig,
	}

	rundownNodeRenderer := renderer.NewRundownNodeRenderer()

	ar := ansi.NewRenderer(ansiOptions)
	r := goldrenderer.NewRenderer(
		goldrenderer.WithNodeRenderers(
			goldutil.Prioritized(ar, 1000),
			goldutil.Prioritized(rundownNodeRenderer, 1000),
		),
	)

	rundownRenderer := renderer.NewRundownRenderer(
		r,
	)

	gm := goldmark.New(
		goldmark.WithParserOptions(
			parser.WithASTTransformers(goldutil.PrioritizedValue{
				Value:    transformer.NewRundownASTTransformer(),
				Priority: 0,
			}),
		),
		goldmark.WithRenderer(rundownRenderer),
	)

	data, _ := ioutil.ReadFile(rundownFile)
	doc := gm.Parser().Parse(text.NewReader(data))

	for child := doc.FirstChild(); child != nil; child = child.NextSibling() {
		if section, ok := child.(*ast.SectionPointer); ok {
			cmd := ports.BuildCobraCommand(rundownFile, section)
			if cmd != nil {
				docRoot.AddCommand(cmd)
			}
		}
	}

	// rd, err := rundown.LoadFile(rundownFile)
	// if err != nil {
	// 	return docRoot // If we can't find a file, just return the rundown command itself.
	// }

	// rd.SetLogger(flagDebug)

	// shortCodes := rd.GetShortCodes()

	// addFlags(docRoot, shortCodes.Options)

	// for _, sc := range shortCodes.Codes {
	// 	shortCode := sc
	// 	reqOpts := []string{}

	// 	for _, opt := range shortCode.Options {
	// 		if opt.Required {
	// 			reqOpts = append(reqOpts, fmt.Sprintf("--%s %s", opt.Code, strings.ToUpper(opt.Type)))
	// 		}
	// 	}

	// 	codeCommand := &cobra.Command{
	// 		Use:   fmt.Sprintf("%s [flags]", shortCode.Code),
	// 		Short: shortCode.Name,
	// 		Long:  fmt.Sprintf("%s is a command within %s.\n\n%s", shortCode.Code, relRundownFile, shortCode.Description),
	// 		RunE: func(cmd *cobra.Command, args []string) error {
	// 			docSpec, err := buildDocSpec(shortCode.Code, rd, cmd, args)

	// 			// fmt.Printf("Build docSpec %s\n", docSpec)

	// 			if err == nil {
	// 				err, callback := rd.RunCodes(docSpec)

	// 				if err != nil {
	// 					util.Debugf("ERROR %#v\n", err)

	// 					if errors.Is(err, rundown.InvocationError) {
	// 						cmd.HelpFunc()(cmd, args)
	// 					}
	// 				}

	// 				return handleError(cmd.OutOrStderr(), err, callback)
	// 			}

	// 			return err
	// 		},
	// 	}

	// 	addArgs(codeCommand, shortCode.Options)
	// 	addFlags(codeCommand, shortCode.Options)
	// 	addFlags(codeCommand, shortCodes.Options) // Add doc options

	// 	docRoot.AddCommand(codeCommand)
	// }

	// wd, err := os.Getwd()
	// if err != nil {
	// 	wd = "/"
	// }

	// rdf, err := filepath.Rel(wd, rundownFile)
	// if err != nil {
	// 	rdf = rundownFile
	// }

	// docRoot.Long = fmt.Sprintf("%s\n\nCurrent file: %s\n\n", docRoot.Long, rdf)

	return docRoot
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "rundown [command] [flags]...",
		Short:         "Execute a markdown file",
		Long:          `Rundown turns Markdown files into console scripts.`,
		SilenceUsage:  true,
		SilenceErrors: false,
		PreRun: func(cmd *cobra.Command, args []string) {
			rundownFile = shared.RundownFile(flagFilename)

			util.Debugf("RundownFile = %s\n", rundownFile)

			if len(rundownFile) == 0 {
				if len(flagFilename) == 0 {
					println("Could not find RUNDOWN.md or README.md in current or parent directories.")
				} else {
					println("Could not read file ", flagFilename)
				}

				os.Exit(1)
			}

			if len(args) > 0 {
				argShortcodes = args
			} else if flagDefault != "" {
				argShortcodes = []string{flagDefault}
			} else {
				argShortcodes = []string{}
			}

		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args)
		},
	}

	rootCmd.PersistentFlags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")
	rootCmd.PersistentFlags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.PersistentFlags().BoolVarP(&flagViewOnly, "display", "d", false, "Render without executing scripts")
	rootCmd.PersistentFlags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")

	rootCmd.Flag("cols").Hidden = true
	rootCmd.Flag("display").Hidden = true
	rootCmd.Flag("completions").Hidden = true

	return rootCmd
}

func init() {

}
