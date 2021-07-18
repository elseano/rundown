package cmd

import (
	"io/ioutil"
	"os"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glamour/ansi"
	shared "github.com/elseano/rundown/cmd"
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
	context := renderer.NewContext()
	rundownNodeRenderer := renderer.NewRundownNodeRenderer(context)

	ar := ansi.NewRenderer(ansiOptions)
	r := goldrenderer.NewRenderer(
		goldrenderer.WithNodeRenderers(
			goldutil.Prioritized(ar, 1000),
			goldutil.Prioritized(rundownNodeRenderer, 1000),
		),
	)

	rundownRenderer := renderer.NewRundownRenderer(
		r,
		context,
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
			return nil
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
