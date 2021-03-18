package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "rundown [filename] [shortcodes]...",
	Short: "Execute a markdown file",
	Long:  `Rundown turns Markdown files into applications`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("Must specify at least the filename")
		}

		return nil
	},
	PreRun: func(cmd *cobra.Command, args []string) {
		argFilename = args[0]

		if len(args) > 1 {
			argShortcodes = args[1:]
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
		if len(args) == 0 {
			return []string{"md"}, cobra.ShellCompDirectiveFilterFileExt
		}

		cleanArgs := cmd.Flags().Args()
		cleanArgs = append(cleanArgs, toComplete)
		argFilename = cleanArgs[0]

		completions := performCompletion(cleanArgs)

		if strings.Index(toComplete, "+") == 0 {
			return completions, cobra.ShellCompDirectiveNoSpace
		} else {
			return completions, cobra.ShellCompDirectiveNoFileComp
		}
	},
}

func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute(version string, gitCommit string) error {
	rootCmd.Version = version + " (" + gitCommit + ")"
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&flagCodes, "help", "h", false, "Displays help & shortcodes for the given file")
	rootCmd.Flags().BoolVar(&flagDebug, "debug", false, "Write debugging into to debug.log")
	rootCmd.Flags().BoolVarP(&flagAsk, "ask", "a", false, "Ask which shortcode to run")
	rootCmd.Flags().BoolVar(&flagAskRepeat, "ask-repeat", false, "Continually ask which shortcode to run")
	rootCmd.Flags().StringVar(&flagDefault, "default", "", "Default shortcode to run if none specified")
	rootCmd.Flags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")

	rootCmd.AddCommand(astCmd)
	rootCmd.AddCommand(emojiCmd)
	rootCmd.AddCommand(checkCmd)
	rootCmd.AddCommand(completionCmd)
	rootCmd.AddCommand(viewCmd)

	originalHelpFunc := rootCmd.HelpFunc()

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Populates flags so we can show the default command
		// when we're asking for help from an shebang script.
		rootCmd.ParseFlags(args)

		pureArgs := cmd.Flags().Args()
		if len(pureArgs) > 0 && pureArgs[0] != "" {
			argFilename = pureArgs[0]

			help(cmd, pureArgs)
		} else {
			originalHelpFunc(cmd, args)
		}
	})
}

func help(cmd *cobra.Command, args []string) {
	if argFilename == "" {
		argFilename = args[0]
	}

	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	codes, err := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
	if err == nil {
		rd.RunCodesWithoutValidation(codes)
	}

	err = RenderShortCodes()
	if err != nil {
		fmt.Printf("Document %s provides no help information %s.\n", argFilename, err.Error())
	}

	os.Exit(0)
}

func run(cmd *cobra.Command, args []string) int {
	var callBack func()

	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)
	rd.SetOutput(cmd.OutOrStdout())
	rd.SetConsoleWidth(flagCols)

	switch {

	case flagCodes:

		codes, err := rundown.ParseShortCodeSpecs([]string{"help"})
		if err == nil {
			rd.RunCodes(codes)
		}

		err = RenderShortCodes()
		return handleError(cmd.OutOrStderr(), err, callBack)

	case flagAsk:
		KillReadlineBell()

		spec, err := AskShortCode()
		if err != nil {
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

		if spec != nil {
			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			return handleError(cmd.OutOrStderr(), err, callBack)
		}

	case flagAskRepeat:
		KillReadlineBell()

		for {
			spec, err := AskShortCode()
			if err != nil {
				return handleError(cmd.OutOrStderr(), err, callBack)
			}

			if spec == nil {
				break
			}

			err, callBack = rd.RunCodes(&rundown.DocumentSpec{ShortCodes: []*rundown.ShortCodeSpec{spec}, Options: map[string]*rundown.ShortCodeOptionSpec{}})
			return handleError(cmd.OutOrStderr(), err, callBack)
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
			if errors.Is(err, rundown.InvocationError) {
				codes, err2 := rundown.ParseShortCodeSpecs([]string{"rundown:help"})
				if err2 == nil {
					rd.RunCodesWithoutValidation(codes)
				}

				err2 = RenderShortCodes()
			}
		}

		return handleError(cmd.OutOrStderr(), err, callBack)
	}

	return 0
}
