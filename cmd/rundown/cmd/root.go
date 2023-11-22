package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	shared "github.com/elseano/rundown/cmd"
	rundown "github.com/elseano/rundown/pkg"
	"github.com/elseano/rundown/pkg/ast"
	"github.com/elseano/rundown/pkg/ports"
	"github.com/elseano/rundown/pkg/util"
	"github.com/muesli/reflow/indent"
	"github.com/spf13/cobra"
)

func Execute(version string, gitCommit string) error {
	cmd := NewDocRootCmd(os.Args)
	cmd.Version = version
	if gitCommit != "" {
		cmd.Version += " (" + gitCommit + ")"
	}

	if flagCompletions != "" {
		switch flagCompletions {
		case "bash":
			return cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			return cmd.Root().GenZshCompletion(os.Stdout)
		case "fish":
			return cmd.Root().GenFishCompletion(os.Stdout, true)
		case "powershell":
			return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
		}
	}

	return cmd.Execute()
}

func NewDocRootCmd(args []string) *cobra.Command {
	docRoot := NewRootCmd()
	docRoot.ParseFlags(args)
	docRoot.Root().CompletionOptions.DisableDefaultCmd = true

	rundownFile = shared.RundownFile(flagFilename)
	if rundownFile == "" {
		if flagCompletions == "" {
			fmt.Fprintf(os.Stderr, "Error: No RUNDOWN.md file found in current path or parents.\n\n")
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}

	loaded, err := rundown.Load(rundownFile)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if flagFilename != "" {
				fmt.Fprintf(os.Stderr, "Error: Couldn't load file %s. File doesn't exist.\n\n", flagFilename)
			} else {
				fmt.Fprintf(os.Stderr, "Error: No RUNDOWN.md file found in current path or parents.\n\n")
			}
		} else {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		}
	}

	if loaded != nil {
		for _, section := range loaded.GetSections() {
			if !section.Pointer.Silent {
				cmd := ports.BuildCobraCommand(rundownFile, section, flagDebug)
				if cmd != nil {
					docRoot.AddCommand(cmd)
				}
			}
		}
	}

	return docRoot
}

func NewRootCmd() *cobra.Command {

	rundownFile = shared.RundownFile(flagFilename)

	doc, err := rundown.Load(rundownFile)
	longDesc := ""

	if err == nil {
		if help := ast.GetRootHelp(doc.MasterDocument.Document); help != nil {
			str := strings.Builder{}
			writer := indent.NewWriterPipe(&str, 2, nil)
			doc.MasterDocument.Goldmark.Renderer().Render(writer, doc.MasterDocument.Source, help)
			longDesc = "\n\n" + str.String()
		}
	}

	rootCmd := &cobra.Command{
		Use:           "rundown [command] [flags]...",
		Short:         "Execute a markdown file",
		Long:          "Rundown turns Markdown files into console scripts." + longDesc,
		SilenceUsage:  true,
		SilenceErrors: false,
		PreRun: func(cmd *cobra.Command, args []string) {
			rundownFile = shared.RundownFile(flagFilename)
			if flagDebug {
				devNull, _ := os.Create("rundown.log")
				util.RedirectLogger(devNull)
			}

			util.Debugf("RundownFile = %s\n", rundownFile)

			if len(rundownFile) == 0 {
				if len(flagFilename) == 0 {
					// Error already written.
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

			if flagDump {
				doc, _ := rundown.Load(rundownFile)
				doc.MasterDocument.Document.Dump(doc.MasterDocument.Source, 0)
			}

			return nil
		},
	}

	rootCmd.Flags().BoolVar(&flagDump, "dump", false, "Dump the AST only")

	rootCmd.PersistentFlags().StringVarP(&flagFilename, "file", "f", "", "File to run (defaults to RUNDOWN.md then README.md)")
	rootCmd.PersistentFlags().StringVar(&flagCompletions, "completions", "", "Render shell completions for given shell (bash, zsh, fish, powershell)")
	rootCmd.PersistentFlags().BoolVar(&flagDebug, "debug", false, "Write debugging info to rundown.log")
	rootCmd.PersistentFlags().StringVar(&flagServePort, "serve", "", "Set the port to serve a HTML interface for Rundown")
	rootCmd.PersistentFlags().Bool("dump", false, "Dump the AST to be executed")

	rootCmd.Flag("completions").Hidden = true
	rootCmd.Flag("dump").Hidden = true
	rootCmd.Flag("debug").Hidden = true

	// Serve command not working yet.
	rootCmd.Flag("serve").Hidden = true

	return rootCmd
}

func init() {

}
