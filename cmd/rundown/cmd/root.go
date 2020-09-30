package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/elseano/rundown/internal/cli"
	"github.com/elseano/rundown/pkg/segments"
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
	logger := cli.BuildLogger(flagDebug)

	if flagCodes {
		cli.DisplayShortCodes(argFilename, logger)
	} else if len(argShortcodes) > 0 {
		cli.RunShortCode(segments.NewContext(), argFilename, argShortcodes, logger)
	} else {
		if flagAsk {
			cli.RunHeading(segments.NewContext(), argFilename, logger)
		} else if flagAskRepeat {
			context := segments.NewContext()
			context.Repeat = true
			context.Invocation = strings.Join(os.Args, " ")

			for {
				cli.RunHeading(context, argFilename, logger)
			}
		} else {
			cli.ExecuteFile(argFilename, logger)
		}
	}
	fmt.Printf("\n")
}
