package cmd

import (
	"fmt"
	"os"

	shared "github.com/elseano/rundown/cmd"
	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ports"
	"github.com/elseano/rundown/pkg/util"
	"github.com/spf13/cobra"
)

func Execute(version string, gitCommit string) error {
	cmd := NewDocRootCmd(os.Args)
	cmd.Version = version + " (" + gitCommit + ")"

	return cmd.Execute()
}

func NewDocRootCmd(args []string) *cobra.Command {
	docRoot := NewRootCmd()
	docRoot.ParseFlags(args)

	rundownFile = shared.RundownFile(flagFilename)

	loaded, err := rundown.Load(rundownFile)
	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}

	for _, section := range loaded.GetSections() {
		cmd := ports.BuildCobraCommand(rundownFile, section, flagDebug)
		if cmd != nil {
			docRoot.AddCommand(cmd)
		}
	}

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
			if flagServePort != "" {
				ports.ServeRundown(rundownFile, flagDebug, flagServePort)
			}

			return nil
		},
	}

	rootCmd.PersistentFlags().IntVar(&flagCols, "cols", util.IntMin(util.GetConsoleWidth(), 120), "Number of columns in display")
	rootCmd.PersistentFlags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.PersistentFlags().BoolVarP(&flagViewOnly, "display", "d", false, "Render without executing scripts")
	rootCmd.PersistentFlags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Write debugging info to rundown.log")
	rootCmd.PersistentFlags().StringVar(&flagServePort, "serve", "", "Set the port to serve a HTML interface for Rundown")
	rootCmd.PersistentFlags().Bool("dump", false, "Dump the AST to be executed")

	rootCmd.Flag("cols").Hidden = true
	rootCmd.Flag("display").Hidden = true
	rootCmd.Flag("completions").Hidden = true

	return rootCmd
}

func init() {

}
