package cmd

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rundown [filename] [shortcodes]...",
	Short: "Execute a markdown file",
	Long:  `Rundown turns Markdown files into applications`,
	Args: func(cmd *cobra.Command, args []string) error {
		if findMarkdownFile(flagFilename) == "" {
			return errors.New("No RUNDOWN.md or README.md file found. Use -f to specify a filename.")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		rundownFile = findMarkdownFile(flagFilename)

		if len(args) > 0 {
			argShortcodes = args
		} else if flagDefault != "" {
			argShortcodes = []string{flagDefault}
		}

	},
	Run: func(cmd *cobra.Command, args []string) {
		exitCode := run(cmd, args)
		os.Exit(exitCode)
	},
	ValidArgs: []string{},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cleanArgs := cmd.Flags().Args()
		cleanArgs = append(cleanArgs, toComplete)

		rundownFile = findMarkdownFile(flagFilename)

		completions := performCompletion(cleanArgs)

		if strings.Index(toComplete, "+") == 0 {
			return completions, cobra.ShellCompDirectiveNoSpace
		} else {
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
	},
}

func findMarkdownFile(file string) string {
	dir, err := os.Getwd()

	if len(file) != 0 {
		return file
	}

	if err != nil {
		util.Debugf("Err: %#v\n", err)
		if util.FileExists("RUNDOWN.md") {
			return "RUNDOWN.md"
		} else {
			return "README.md"
		}
	}

	// Search the CWD and parent paths until the root path for RUNDOWN.md/README.md
	for {
		for _, fn := range []string{"RUNDOWN.md", "README.md"} {
			path := path.Join(dir, fn)
			util.Debugf("Searching: %s\n", path)

			if util.FileExists(path) {
				return path
			}
		}

		if dir == "/" {
			return "README.md" // Will fail out later.
		}

		dir = path.Dir(dir)
	}
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute(version string, gitCommit string) error {
	rootCmd.Version = version + " (" + gitCommit + ")"

	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVar(&flagDebug, "debug", false, "Write debugging into to debug.log")
	rootCmd.Flags().BoolVarP(&flagAsk, "ask", "a", false, "Ask which shortcode to run")
	rootCmd.Flags().BoolVar(&flagAskRepeat, "ask-repeat", false, "Continually ask which shortcode to run")
	rootCmd.Flags().StringVar(&flagDefault, "default", "", "Default shortcode to run if none specified")
	rootCmd.Flags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")
	rootCmd.Flags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.Flags().BoolVarP(&flagViewOnly, "display", "d", false, "Render without executing scripts")
	rootCmd.Flags().BoolVarP(&flagCheckOnly, "check", "c", false, "Check file is valid for Rundown")
	rootCmd.Flags().BoolVarP(&flagAst, "ast", "i", false, "Inspect the Rundown AST")
	rootCmd.Flags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")

	originalHelpFunc := rootCmd.HelpFunc()

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Populates flags so we can show the default command
		// when we're asking for help from an shebang script.
		rootCmd.ParseFlags(args)

		pureArgs := cmd.Flags().Args()
		originalHelpFunc(cmd, args)

		// Set rundown file, as root's PreRun doesn't get run for help.
		rundownFile = findMarkdownFile(flagFilename)

		if rundownFile != "" {
			help(cmd, pureArgs)
		}
	})
}

func help(cmd *cobra.Command, args []string) {
	rd, err := rundown.LoadFile(rundownFile)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	cleanFilename := filepath.Base(rundownFile)

	fmt.Printf("\nHelp for %s\n\n", cleanFilename)

	d, _ := rd.GetAST()
	doc := d.(*markdown.SectionedDocument)
	for desc := doc.Sections[0].Description.Front(); desc != nil; desc = desc.Next() {
		if text := desc.Value.(*markdown.RundownBlock).GetModifiers().GetValue("desc"); text != nil {
			fmt.Printf("  %s\n", *text)
		}
	}

	codes, err := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
	if err == nil {
		rd.RunCodesWithoutValidation(codes)
	}

	err = RenderShortCodes()
	if err != nil {
		fmt.Printf("Document %s provides no help information %s.\n", rundownFile, err.Error())
	}

	os.Exit(0)
}

func run(cmd *cobra.Command, args []string) int {
	var callBack func()

	rd, err := rundown.LoadFile(findMarkdownFile(flagFilename))
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	KillReadlineBell()

	switch {

	case flagAst:

		return runAst(rundownFile)

	case flagCheckOnly:

		return runCheck(rundownFile)

	case flagCompletions != "":

		runCompletions(flagCompletions, cmd)
		return 0

	case flagCodes:

		codes, err := rundown.ParseShortCodeSpecs([]string{"help"})
		if err == nil {
			rd.RunCodes(codes)
		}

		err = RenderShortCodes()
		return handleError(cmd.OutOrStderr(), err, callBack)

	case flagAsk:
		spec, err := AskShortCode()
		if err != nil {
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

		if spec != nil {
			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

	case flagAskRepeat:
		for {
			spec, err := AskShortCode()
			if err != nil {
				return handleError(cmd.OutOrStderr(), err, callBack)
			}

			if spec == nil {
				break
			}

			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			handleError(cmd.OutOrStderr(), err, callBack)
		}

	default:
		specs := argShortcodes

		if len(specs) == 0 && flagDefault != "" {
			specs = []string{flagDefault}
		}

		codes, err := rundown.ParseShortCodeSpecs(specs)
		if err == nil {
			err, callBack = rd.RunCodes(codes)
		}

		if err != nil {
			util.Debugf("ERROR %#v\n", err)
			if errors.Is(err, rundown.InvocationError) {
				codes, err2 := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
				if err2 == nil {
					rd.RunCodesWithoutValidation(codes)
				}

				util.Debugf("Rendering shortcodes for %s\n", rd.Filename())
				err2 = RenderShortCodes()
				util.Debugf("ERROR %#v\n", err2)
			}
		}

		return handleError(cmd.OutOrStderr(), err, callBack)
	}

	return 0
}
