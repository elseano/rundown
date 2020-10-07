package cmd

import (
	"errors"

	"github.com/elseano/rundown/pkg/rundown"
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
	Run: run,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Version = Version + " (" + GitCommit + ")"

	rootCmd.Flags().BoolVarP(&flagCodes, "codes", "c", false, "Displays available shortcodes for the given file")
	rootCmd.Flags().BoolVar(&flagDebug, "debug", false, "Write debugging into to debug.log")
	rootCmd.Flags().BoolVarP(&flagAsk, "ask", "a", false, "Ask which shortcode to run")
	rootCmd.Flags().BoolVar(&flagAskRepeat, "ask-repeat", false, "Continually ask which shortcode to run")
	rootCmd.Flags().StringVar(&flagDefault, "default", "", "Default shortcode to run if none specified")

	rootCmd.AddCommand(astCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(emojiCmd)
	rootCmd.AddCommand(checkCmd)
}

func run(cmd *cobra.Command, args []string) {
	rd, err := rundown.LoadFile(argFilename)
	if err != nil {
		panic(err)
	}

	rd.SetLogger(flagDebug)

	switch {

	case flagCodes:
		err := RenderShortCodes()
		handleError(err)

	case flagAsk:
		KillReadlineBell()

		spec, err := AskShortCode()
		handleError(err)

		if spec != nil {
			err = rd.RunCodes([]*rundown.ShortCodeSpec{spec})
			handleError(err)
		}

	case flagAskRepeat:
		KillReadlineBell()

		for {
			spec, err := AskShortCode()
			handleError(err)

			if spec == nil {
				break
			}

			err = rd.RunCodes([]*rundown.ShortCodeSpec{spec})
			handleError(err)
		}

	case len(argShortcodes) > 0 || flagDefault != "":
		specs := argShortcodes

		if len(specs) == 0 {
			specs = []string{flagDefault}
		}

		codes, err := rundown.ParseShortCodeSpecs(specs)
		handleError(err)

		err = rd.RunCodes(codes)
		handleError(err)

	default:
		err = rd.RunSequential()
		handleError(err)

	}

}
