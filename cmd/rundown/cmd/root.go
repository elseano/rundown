package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	shared "github.com/elseano/rundown/cmd"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

func Execute(version string, gitCommit string) error {
	cmd := NewDocRootCmd(os.Args)
	cmd.Version = version + " (" + gitCommit + ")"

	return cmd.Execute()
}

func addFlags(cmd *cobra.Command, options map[string]*rundown.ShortCodeOption) {
	for _, opt := range options {
		switch opt.Type {
		case "string":
			cmd.Flags().String(opt.Code, opt.Default, opt.Description)
		case "file-exists":
			cmd.Flags().String(opt.Code, opt.Default, opt.Description)
			cmd.MarkFlagFilename(opt.Code)
		case "file-not-exists":
			cmd.Flags().String(opt.Code, opt.Default, opt.Description)
		case "bool":
			cmd.Flags().Bool(opt.Code, false, opt.Description)
		case "int":
			cmd.Flags().Int32(opt.Code, -1, opt.Description)
		default:
			cmd.Flags().String(opt.Code, opt.Default, opt.Description)
		}

		// cmd.MarkFlagRequired(opt.Code)
	}
}

func NewDocRootCmd(args []string) *cobra.Command {
	docRoot := NewRootCmd()
	docRoot.ParseFlags(args)

	rundownFile = shared.RundownFile(flagFilename)

	rd, err := rundown.LoadFile(rundownFile)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	shortCodes := rd.GetShortCodes()

	addFlags(docRoot, shortCodes.Options)

	for _, shortCode := range shortCodes.Codes {
		reqOpts := []string{}

		for _, opt := range shortCode.Options {
			if opt.Required {
				reqOpts = append(reqOpts, fmt.Sprintf("--%s %s", opt.Code, strings.ToUpper(opt.Type)))
			}
		}

		codeCommand := &cobra.Command{
			Use:   fmt.Sprintf("%s %s", shortCode.Code, strings.Join(reqOpts, " ")),
			Short: shortCode.Name,
			Long:  shortCode.Description,
			RunE: func(cmd *cobra.Command, args []string) error {
				docSpec, err := buildDocSpec(shortCode.Code, rd, cmd, args)

				fmt.Printf("Build docSpec %#v\n", docSpec)

				if err == nil {
					err, _ := rd.RunCodes(docSpec)
					return err
				}

				return err
			},
		}

		addFlags(codeCommand, shortCode.Options)
		addFlags(codeCommand, shortCodes.Options) // Add doc options

		docRoot.AddCommand(codeCommand)
	}

	wd, err := os.Getwd()
	if err != nil {
		wd = "/"
	}

	rdf, err := filepath.Rel(wd, rundownFile)
	if err != nil {
		rdf = rundownFile
	}

	docRoot.Long = fmt.Sprintf("%s\n\nCurrent file: %s\n\n", docRoot.Long, rdf)

	return docRoot
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "rundown [flags] [command] [flags]...",
		Short:         "Execute a markdown file",
		Long:          `Rundown turns Markdown files into console scripts.`,
		SilenceUsage:  false,
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

				os.Exit(-1)
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
		ValidArgs: []string{},
		// ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		// 	cleanArgs := cmd.Flags().Args()
		// 	cleanArgs = append(cleanArgs, toComplete)

		// 	rundownFile = shared.RundownFile(flagFilename)

		// 	completions := performCompletion(cleanArgs)

		// 	if strings.Index(toComplete, "+") == 0 {
		// 		return completions, cobra.ShellCompDirectiveNoSpace
		// 	} else {
		// 		return completions, cobra.ShellCompDirectiveNoFileComp
		// 	}
		// },
	}

	// rootCmd.Flags().BoolVar(&flagDebug, "debug", false, "Write debugging into to debug.log")
	// rootCmd.Flags().BoolVarP(&flagAsk, "", "a", false, "Ask which shortcode to run")
	// rootCmd.Flags().BoolVar(&flagAskRepeat, "ask-repeat", false, "Continually ask which shortcode to run")
	// rootCmd.Flags().StringVar(&flagDefault, "default", "", "Default shortcode to run if none specified")
	rootCmd.Flags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")
	rootCmd.Flags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.Flags().BoolVarP(&flagViewOnly, "display", "d", false, "Render without executing scripts")
	rootCmd.Flags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")

	rootCmd.Flag("cols").Hidden = true
	rootCmd.Flag("display").Hidden = true
	rootCmd.Flag("completions").Hidden = true

	// originalHelpFunc := rootCmd.HelpFunc()

	// rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
	// 	// Populates flags so we can show the default command
	// 	// when we're asking for help from an shebang script.
	// 	rootCmd.ParseFlags(args)

	// 	pureArgs := cmd.Flags().Args()
	// 	originalHelpFunc(cmd, args)

	// 	// Set rundown file, as root's PreRun doesn't get run for help.
	// 	rundownFile = shared.RundownFile(flagFilename)

	// 	if rundownFile != "" {
	// 		help(cmd, pureArgs)
	// 	}
	// })

	return rootCmd

}

func init() {

}
