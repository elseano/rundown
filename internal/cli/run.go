package cli

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/markdown"
	"github.com/elseano/rundown/pkg/rundown"
	"github.com/elseano/rundown/pkg/segments"
	"github.com/logrusorgru/aurora"
	"github.com/urfave/cli/v2"
	"github.com/yuin/goldmark/text"
	"golang.org/x/tools/godoc/util"
)

func BuildLogger(debugging bool) *log.Logger {
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

func FileToSegments(filename string, logger *log.Logger) []segments.Segment {
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

func GetShortCodes(filename string, logger *log.Logger) []string {
	loadedSegments := FileToSegments(filename, logger)
	codes := []string{}

	for _, segment := range loadedSegments {
		if heading, ok := segment.(*segments.HeadingMarker); ok && heading.ShortCode != "" {
			codes = append(codes, heading.ShortCode)
		}
	}

	return codes
}

func DisplayShortCodes(filename string, logger *log.Logger) {
	loadedSegments := FileToSegments(filename, logger)

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

func RunShortCode(context *rundown.Context, filename string, requestedShortCodes []string, logger *log.Logger) {
	loadedSegments := FileToSegments(filename, logger)

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

	result := segments.ExecuteRundown(rundown.NewContext(), toRun, md.Renderer(), logger, os.Stdout)
	if result.IsError {
		os.Exit(1)
	}

}

func RunHeading(context *rundown.Context, filename string, logger *log.Logger) {
	loadedSegments := FileToSegments(filename, logger)

	fmt.Printf(aurora.Faint("Rundown running %s\n\n").String(), filename)

	shortCode := ShortcodeMenu(loadedSegments)

	fmt.Println()

	if shortCode == "" {
		os.Exit(0)
	}

	RunShortCode(context, filename, []string{shortCode}, logger)

	fmt.Println()
}

func ExecuteFile(filename string, logger *log.Logger) {
	loadedSegments := FileToSegments(filename, logger)
	result := segments.ExecuteRundown(rundown.NewContext(), loadedSegments, md.Renderer(), logger, os.Stdout)
	if result.IsError {
		os.Exit(1)
	}
}

func InspectRundown(c *cli.Context) {
	logger := BuildLogger(c.Bool("debug"))
	filename := c.Args().Get(0)

	loadedSegments := FileToSegments(filename, logger)

	for _, x := range loadedSegments {
		fmt.Println(x.String())
	}

}

func InspectMarkdown(c *cli.Context) {
	filename := c.Args().Get(0)

	md := markdown.PrepareMarkdown()

	b, _ := ioutil.ReadFile(filename)

	reader := text.NewReader(b)

	doc := md.Parser().Parse(reader)

	doc.Dump(b, 0)
}

func DefaultRun(c *cli.Context) error {
	logger := BuildLogger(c.Bool("debug"))
	shortCode := []string{}
	filename := c.Args().Get(0)

	if c.Args().Len() > 1 {
		for i := 1; i < c.Args().Len(); i = i + 1 {
			shortCode = append(shortCode, c.Args().Get(i))
		}
	} else if c.String("default") != "" {
		shortCode = []string{c.String("default")}
	}

	if c.Bool("codes") {
		DisplayShortCodes(filename, logger)
	} else if len(shortCode) > 0 {
		RunShortCode(rundown.NewContext(), filename, shortCode, logger)
	} else {
		if c.Bool("ask") {
			RunHeading(rundown.NewContext(), filename, logger)
		} else if c.Bool("ask-repeat") {
			context := rundown.NewContext()
			context.Repeat = true
			context.Invocation = strings.Join(os.Args, " ")

			for {
				RunHeading(context, filename, logger)
			}
		} else {
			ExecuteFile(filename, logger)
		}
	}
	fmt.Printf("\n")

	return nil
}

func DefaultComplete(c *cli.Context) {
	if c.Args().Len() > 0 {
		logger := BuildLogger(false)
		for _, code := range GetShortCodes(c.Args().Get(0), logger) {
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
