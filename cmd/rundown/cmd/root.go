package cmd

import (
	"os"
	"strings"

	shared "github.com/elseano/rundown/cmd"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

var rootCmd = NewRootCmd()

func RootCmd() *cobra.Command {
	return rootCmd
}

func Execute(version string, gitCommit string) error {
	rootCmd.Version = version + " (" + gitCommit + ")"

	return rootCmd.Execute()
}

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:           "rundown [filename] [shortcodes]...",
		Short:         "Execute a markdown file",
		Long:          `Rundown turns Markdown files into applications`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			rundownFile = shared.RundownFile(flagFilename)

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
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			cleanArgs := cmd.Flags().Args()
			cleanArgs = append(cleanArgs, toComplete)

			rundownFile = shared.RundownFile(flagFilename)

			completions := performCompletion(cleanArgs)

			if strings.Index(toComplete, "+") == 0 {
				return completions, cobra.ShellCompDirectiveNoSpace
			} else {
				return completions, cobra.ShellCompDirectiveNoFileComp
			}
		},
	}

	rootCmd.Flags().BoolVar(&flagDebug, "debug", false, "Write debugging into to debug.log")
	rootCmd.Flags().BoolVarP(&flagAsk, "ask", "a", false, "Ask which shortcode to run")
	rootCmd.Flags().BoolVar(&flagAskRepeat, "ask-repeat", false, "Continually ask which shortcode to run")
	rootCmd.Flags().StringVar(&flagDefault, "default", "", "Default shortcode to run if none specified")
	rootCmd.Flags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")
	rootCmd.Flags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.Flags().BoolVarP(&flagViewOnly, "display", "d", false, "Render without executing scripts")
	rootCmd.Flags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")

	originalHelpFunc := rootCmd.HelpFunc()

	rootCmd.SetHelpFunc(func(cmd *cobra.Command, args []string) {
		// Populates flags so we can show the default command
		// when we're asking for help from an shebang script.
		rootCmd.ParseFlags(args)

		pureArgs := cmd.Flags().Args()
		originalHelpFunc(cmd, args)

		// Set rundown file, as root's PreRun doesn't get run for help.
		rundownFile = shared.RundownFile(flagFilename)

		if rundownFile != "" {
			help(cmd, pureArgs)
		}
	})

	return rootCmd

}

func init() {

}
