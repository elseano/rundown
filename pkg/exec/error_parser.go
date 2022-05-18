package exec

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/elseano/rundown/pkg/exec/scripts"
	"github.com/elseano/rundown/pkg/util"
	"github.com/logrusorgru/aurora"
)

type ErrorSource struct {
	Source []byte
	Line   int
}

type ErrorDetails struct {
	Error       string
	ErrorSource *ErrorSource
}

func ParseError(script *scripts.Script, stdout string) *ErrorDetails {
	if index := strings.Index(stdout, script.AbsolutePath); index != -1 {
		return parseError(stdout, index, script)
	}

	return &ErrorDetails{
		Error: stdout,
	}
}

var lineMatcher = regexp.MustCompile(`[A-Za-z0-9_/\/-]+\:((\d+)| line\:? *(\d+))\: *(.*)`)

func parseError(stdout string, errorIndex int, source *scripts.Script) *ErrorDetails {
	util.Logger.Debug().Fields(map[string]interface{}{"STDOUT": stdout}).Msgf("Looking up error in STDOUT")

	if matches := lineMatcher.FindStringSubmatch(stdout[errorIndex:]); matches != nil {
		line := 0
		if matches[2] != "" {
			line, _ = strconv.Atoi(matches[2])
		} else {
			line, _ = strconv.Atoi(matches[3])
		}

		line -= lineOffset(source)

		errText := matches[4]

		return &ErrorDetails{
			Error: errText,
			ErrorSource: &ErrorSource{
				Source: source.OriginalContents,
				Line:   line,
			},
		}
	}

	return &ErrorDetails{Error: stdout}
}

func (e *ErrorDetails) String(colors aurora.Aurora) string {
	output := strings.Builder{}

	if e.ErrorSource != nil {
		lines := strings.Split(strings.ReplaceAll(string(e.ErrorSource.Source), "\r\n", "\n"), "\n")
		lastLine := len(lines) - 1
		for i, line := range lines {

			if i == lastLine && line == "" {
				continue
			}

			lineIndicator := " "
			if e.ErrorSource.Line == i {
				lineIndicator = colors.Red("*").String()
			}

			output.WriteString(fmt.Sprintf("%s %s %s\n", lineIndicator, fmt.Sprintf(colors.Faint("%4d:").String(), i+1), line))
		}

		output.WriteString("\n")
		output.WriteString(colors.Sprintf(colors.Faint("Line %d: "), e.ErrorSource.Line))
	}

	output.WriteString(fmt.Sprintf("%s\n", e.Error))

	return output.String()
}

func lineOffset(script *scripts.Script) int {
	if script.Prefix == nil {
		return 0
	}

	return bytes.Count(script.Prefix, []byte("\n"))
}
