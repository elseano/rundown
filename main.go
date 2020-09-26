package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	rdcli "github.com/elseano/rundown/cli"
	"github.com/elseano/rundown/markdown"
	"github.com/elseano/rundown/segments"

	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
	"golang.org/x/tools/godoc/util"
)

func buildLogger(debugging bool) *log.Logger {
	var debug io.Writer

	if debugging {
		debug, _ = os.Create("debug.log")
	} else {
		debug = ioutil.Discard
	}

	logger := log.New(debug, "", log.Ltime)

	return logger
}

var md = markdown.PrepareMarkdown()

func fileToSegments(filename string, logger *log.Logger) []segments.Segment {
	logger.Printf("Loading file %s", filename)

	b, err := ioutil.ReadFile(filename)

	// Trim shebang
	if bytes.HasPrefix(b, []byte("#!")) {
		b2 := bytes.SplitN(b, []byte("\n"), 2)
		if len(b2) == 2 {
			b = b2[1]
		}
	}

	if err != nil {
		if os.IsNotExist(err) {
			fmt.Printf("File not found %s\n", filename)
			os.Exit(1)
		} else {
			panic(err)
		}
	}

	result := segments.BuildSegments(string(b), md, logger)

	return result
}

func getShortCodes(filename string, logger *log.Logger) []string {
	loadedSegments := fileToSegments(filename, logger)
	codes := []string{}

	for _, segment := range loadedSegments {
		if heading, ok := segment.(*segments.HeadingMarker); ok && heading.ShortCode != "" {
			codes = append(codes, heading.ShortCode)
		}
	}

	return codes
}

func displayShortCodes(filename string, logger *log.Logger) {
	loadedSegments := fileToSegments(filename, logger)

	sections := []*segments.HeadingMarker{}
	longestHeading := 0

	for _, segment := range loadedSegments {
		if heading, ok := segment.(*segments.HeadingMarker); ok && heading.ShortCode != "" {
			sections = append(sections, heading)
			if longestHeading < len(heading.ShortCode) {
				longestHeading = len(heading.ShortCode)
			}
		}
	}

	fmtStr := "  %-" + strconv.Itoa(longestHeading+2) + "s    %s %s\n"

	fmt.Printf("This file supports jumping to specific sections.\n")
	fmt.Printf("\nUsage:\n")
	fmt.Printf("    rundown %s [SHORTCODE]\n", filename)
	fmt.Printf("\nWith the following short codes availables:\n")

	for _, heading := range sections {
		var desc = ""

		if heading.Description != "" {
			desc = "- " + heading.Description
		}
		fmt.Printf(fmtStr, heading.ShortCode, heading.Title, desc)
	}

	os.Exit(0)
}

func runShortCode(context *segments.Context, filename string, requestedShortCodes []string, logger *log.Logger) {
	loadedSegments := fileToSegments(filename, logger)

	runAt := -1
	toRun := []segments.Segment{}

	for _, requestedShortCode := range requestedShortCodes {

		for i := 0; i < len(loadedSegments); i = i + 1 {
			if heading, ok := loadedSegments[i].(*segments.HeadingMarker); ok && heading.ShortCode == requestedShortCode {
				runAt = i
				break
			}
		}

		if runAt == -1 {
			fmt.Printf("Shortcode %s not found. Run with --codes to see available shortcodes.\n", requestedShortCode)
			os.Exit(1)
		}

		baseLevel := loadedSegments[runAt].GetLevel() - 1

		loadedSegments[runAt].DeLevel(baseLevel)

		toRun = append(toRun, loadedSegments[runAt])
		for i := runAt + 1; i < len(loadedSegments) && !(loadedSegments[i].Kind() == "HeadingMarker" && loadedSegments[i].GetLevel()-1 <= baseLevel); i = i + 1 {
			loadedSegments[i].DeLevel(baseLevel)
			toRun = append(toRun, loadedSegments[i])
		}

		runAt = -1
	}

	fmt.Printf(aurora.Faint("Rundown running %s shortcodes: %s\n\n").String(), filename, strings.Join(requestedShortCodes, ", "))

	result := segments.ExecuteRundown(segments.NewContext(), toRun, md.Renderer(), logger, os.Stdout)
	if result.IsError {
		os.Exit(1)
	}

}

func runHeading(context *segments.Context, filename string, logger *log.Logger) {
	loadedSegments := fileToSegments(filename, logger)

	fmt.Printf(aurora.Faint("Rundown running %s\n\n").String(), filename)

	shortCode := rdcli.ShortcodeMenu(loadedSegments)

	fmt.Println()

	if shortCode == "" {
		os.Exit(0)
	}

	runShortCode(context, filename, []string{shortCode}, logger)

	fmt.Println()
}

func executeFile(filename string, logger *log.Logger) {
	loadedSegments := fileToSegments(filename, logger)
	result := segments.ExecuteRundown(segments.NewContext(), loadedSegments, md.Renderer(), logger, os.Stdout)
	if result.IsError {
		os.Exit(1)
	}
}

func inspectRundown(c *cli.Context) {
	logger := buildLogger(c.Bool("debug"))
	filename := c.Args().Get(0)

	loadedSegments := fileToSegments(filename, logger)

	for _, x := range loadedSegments {
		fmt.Println(x.String())
	}

}

func defaultRun(c *cli.Context) error {
	logger := buildLogger(c.Bool("debug"))
	shortCode := []string{}
	filename := c.Args().Get(0)

	if c.Args().Len() > 1 {
		for i := 1; i < c.Args().Len(); i = i + 1 {
			shortCode = append(shortCode, c.Args().Get(i))
		}
	} else if c.String("default") != "" {
		shortCode = []string{c.String("default")}
	}

	if len(shortCode) > 0 {
		runShortCode(segments.NewContext(), filename, shortCode, logger)
	} else {
		if c.Bool("ask") {
			runHeading(segments.NewContext(), filename, logger)
		} else if c.Bool("ask-repeat") {
			context := segments.NewContext()
			context.Repeat = true
			context.Invocation = strings.Join(os.Args, " ")

			for {
				runHeading(context, filename, logger)
			}
		} else {
			executeFile(filename, logger)
		}
	}
	fmt.Printf("\n")

	return nil
}

func defaultComplete(c *cli.Context) {
	if c.Args().Len() > 0 {
		logger := buildLogger(false)
		for _, code := range getShortCodes(c.Args().Get(0), logger) {
			fmt.Fprintf(c.App.Writer, "%s\n", code)
		}
	} else {
		files, err := ioutil.ReadDir(".")
		if err != nil {
			log.Fatal(err)
		}

		for _, file := range files {
			if file.IsDir() {
				fmt.Fprintf(c.App.Writer, "%s/\n", file.Name())
			} else if file, err := os.Open(file.Name()); err == nil {
				buf := make([]byte, 10)

				if n, err := file.Read(buf); err == nil {
					if util.IsText(buf[0:n]) {
						fmt.Fprintf(c.App.Writer, "%s\n", file.Name())
					}
				}

			}
		}
	}

	return
}

func main() {
	app := &cli.App{
		Name:                   "rundown",
		Usage:                  "Display and execute markdown files.",
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
				Usage: "If no shortcode is provided, ask which section to run (used with shebang scripts)",
			},
			&cli.BoolFlag{
				Name:  "ask-repeat",
				Usage: "If no shortcode is provided, enter an ask and run loop (used with shebang scripts)",
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
					inspectRundown(c)
					return nil
				},
			},
			&cli.Command{
				Name:      "run",
				Aliases:   []string{"r"},
				Usage:     "Runs the markdown file from start to end, or only a specific heading if a shortcode is supplied",
				ArgsUsage: "FILENAME [shortcode] ...",

				Action: func(c *cli.Context) error {
					return defaultRun(c)
				},
				BashComplete: func(c *cli.Context) {
					defaultComplete(c)
				},
			},
			&cli.Command{
				Name:      "select",
				Aliases:   []string{"s"},
				Usage:     "Displays a menu allowing selection of which section (and child sections) to run",
				ArgsUsage: "FILENAME",

				Action: func(c *cli.Context) error {
					logger := buildLogger(c.Bool("debug"))
					runHeading(segments.NewContext(), c.Args().Get(0), logger)
					fmt.Printf("\n")

					return nil
				},
			},
			&cli.Command{
				Name:      "show-codes",
				Aliases:   []string{"c"},
				Usage:     "Displays available shortcodes in the markdown file",
				ArgsUsage: "FILENAME",

				Action: func(c *cli.Context) error {
					logger := buildLogger(c.Bool("debug"))
					displayShortCodes(c.Args().Get(0), logger)
					fmt.Printf("\n")

					return nil
				},
			},
		},
		Action: func(c *cli.Context) error {
			return defaultRun(c)
		},
		BashComplete: func(c *cli.Context) {
			defaultComplete(c)
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
