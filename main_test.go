package main

import (
	"bufio"
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/elseano/rundown/markdown"
	"github.com/elseano/rundown/segments"
	"github.com/elseano/rundown/testutil"
	"github.com/elseano/rundown/util"
)

func TestSimpleRundown(t *testing.T) {
	source, expected := loadFile("_testdata/simple.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestSpacingRundown(t *testing.T) {
	source, expected := loadFile("_testdata/spacing.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestRpcRundown(t *testing.T) {
	source, expected := loadFile("_testdata/rpc.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestStdoutRundown(t *testing.T) {
	source, expected := loadFile("_testdata/stdout.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestFailureRundown(t *testing.T) {
	source, expected := loadFile("_testdata/failure.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestEmojiRundown(t *testing.T) {
	source, expected := loadFile("_testdata/emoji.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestHiddenRundown(t *testing.T) {
	source, expected := loadFile("_testdata/hidden.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestFormattingRundown(t *testing.T) {
	source, expected := loadFile("_testdata/formatting.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestStopOkRundown(t *testing.T) {
	source, expected := loadFile("_testdata/stop.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

func TestStopFailRundown(t *testing.T) {
	source, expected := loadFile("_testdata/stopfail.md")
	actual := run(t, source)

	testutil.AssertLines(t, expected, actual)
}

// Blah
func run(t *testing.T, source string) string {
	logger := testutil.NewTestLogger(t)

	md := markdown.PrepareMarkdown()

	result := segments.BuildSegments(source, md, logger)

	for _, seg := range result {
		t.Log(seg.String())
	}

	var buffer bytes.Buffer

	context := segments.NewContext()
	context.ConsoleWidth = 80

	segments.ExecuteRundown(context, result, md.Renderer(), logger, &buffer)

	fixed := strings.TrimSpace(util.CollapseReturns(util.RemoveColors(buffer.String())))

	for i, line := range strings.Split(fixed, "\n") {
		t.Logf("%3d: %s\n", i, line)
	}

	return fixed
}

func loadFile(filename string) (source string, expected string) {
	fp, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	scanningSource := true

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "-----") {
			scanningSource = false
		} else if scanningSource {
			source = source + "\n" + line
		} else {
			expected = expected + "\n" + line
		}
	}

	return strings.TrimSpace(source), strings.TrimSpace(expected)
}
