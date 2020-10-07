package cmd

import (
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/logrusorgru/aurora"
	"github.com/olekukonko/tablewriter"
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

	if flagCodes {
		shortcodes := rd.GetShortCodes()

		table := tablewriter.NewWriter(os.Stdout)
		table.SetColumnAlignment([]int{tablewriter.ALIGN_RIGHT, tablewriter.ALIGN_LEFT, tablewriter.ALIGN_LEFT})
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetRowLine(false)
		table.SetHeaderLine(false)
		table.SetBorder(false)
		table.SetAutoWrapText(false)

		list := sort.StringSlice{}

		for _, code := range shortcodes {
			list = append(list, code.Code)
		}

		list.Sort()

		for _, codeName := range list {
			code := shortcodes[codeName]

			display := aurora.Bold(code.Name).String()
			if code.Description != "" {
				display = display + "\n" + code.Description
			}

			table.Append([]string{aurora.Bold(code.Code).String(), "", display})

			sortedOptions := sort.StringSlice{}

			for k := range code.Options {
				sortedOptions = append(sortedOptions, k)
			}

			sortedOptions.Sort()

			for _, optCode := range sortedOptions {
				opt := code.Options[optCode]
				spec := ""

				if opt.Default != "" {
					spec = spec + " (default: " + opt.Default + ")"
				} else if opt.Required {
					spec = spec + " (required)"
				}

				table.Append([]string{"", "+" + opt.Code + "=[" + opt.Type + "]", opt.Description + spec})
			}

			table.Append([]string{"", "", ""})
		}

		table.Render()
	} else if len(argShortcodes) > 0 {
		codes, err := rundown.ParseShortCodeSpecs(argShortcodes)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}

		err = rd.RunCodes(codes)
		if err != nil {
			handleError(err)
		}
	} else {
		err = rd.RunSequential()
		if err != nil {
			handleError(err)
		}
	}
}

func handleError(err error) {
	if stopError, ok := err.(*rundown.StopError); ok {
		if stopError.Result.IsError {
			fmt.Printf("\n\n%s - %s in:\n\n", aurora.Bold("Error"), stopError.Result.Message)
			for i, line := range strings.Split(strings.TrimSpace(stopError.Result.Source), "\n") {
				fmt.Printf(aurora.Faint("%3d:").String()+" %s\n", i+1, line)
			}

			fmt.Println()

			fmt.Println(stopError.Result.Output)
			os.Exit(127)
		}
	}

	fmt.Printf("\n\n\n%s: %s\n", aurora.Bold("Error"), err.Error())
	os.Exit(1)
}
