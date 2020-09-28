package main

import (
	"fmt"
	"log"
	"os"

	rdcli "github.com/elseano/rundown/internal/cli"
	"github.com/elseano/rundown/pkg/segments"
	"github.com/urfave/cli/v2"
)

var GitCommit string
var Version string

func main() {
	app := &cli.App{
		Name:                   "rundown",
		Usage:                  "Display and execute markdown files.",
		Version:                Version,
		UsageText:              "rundown [-d] [command] FILENAME shortcode(optional) ...",
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Verbose debugging output to debug.log",
			},
			&cli.StringFlag{
				Name:  "default",
				Usage: "The default shortcode to run if none provided (used with shebang scripts)",
			},
			&cli.BoolFlag{
				Name:  "ask",
				Usage: "If no shortcode is provided, ask which section to run",
			},
			&cli.BoolFlag{
				Name:  "ask-repeat",
				Usage: "If no shortcode is provided, enter an ask and run loop",
			},
			&cli.BoolFlag{
				Name:  "codes",
				Usage: "Lists possible shortcodes in the rundown file",
			},
		},
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			&cli.Command{
				Name:      "inspect",
				Aliases:   []string{"i"},
				Usage:     "Shows the rundown AST of the markdown file",
				ArgsUsage: "FILENAME",

				Action: func(c *cli.Context) error {
					rdcli.InspectRundown(c)
					return nil
				},
			},
			&cli.Command{
				Name:      "ast",
				Usage:     "Shows the markdown AST of the markdown file",
				ArgsUsage: "FILENAME",

				Action: func(c *cli.Context) error {
					rdcli.InspectMarkdown(c)
					return nil
				},
			},
			&cli.Command{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "Runs the markdown file from start to end, or only a specific heading if a shortcode is supplied",
				ArgsUsage: "FILENAME [shortcode] ...",

				Action: func(c *cli.Context) error {
					return rdcli.DefaultRun(c)
				},
				BashComplete: func(c *cli.Context) {
					rdcli.DefaultComplete(c)
				},
			},
			&cli.Command{
				Name:      "select",
				Aliases:   []string{"s"},
				Usage:     "Displays a menu allowing selection of which section (and child sections) to run",
				ArgsUsage: "FILENAME",

				Action: func(c *cli.Context) error {
					logger := rdcli.BuildLogger(c.Bool("debug"))
					rdcli.RunHeading(segments.NewContext(), c.Args().Get(0), logger)
					fmt.Printf("\n")

					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			return rdcli.DefaultRun(c)
		},
		BashComplete: func(c *cli.Context) {
			rdcli.DefaultComplete(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
