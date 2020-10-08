package rundown

import (
	"regexp"
	"strconv"
	"strings"
)

var bashStyle = regexp.MustCompile(": line (\\d+)")
var gccStyle = regexp.MustCompile(":\\s*(\\d+)")

// Attempts to detect the actual line in the script which an error refers to.
func DetectErrorLine(filename string, stdout string) (int, bool) {
	lines := strings.Split(stdout, "\n")

	for i := len(lines) - 1; i >= 0; i-- {
		line := lines[i]
		if filenamePos := strings.Index(line, filename); filenamePos > -1 {
			checkPos := filenamePos + len(filename)
			checkLine := line[checkPos:]

			if match := bashStyle.FindStringSubmatch(checkLine); match != nil {
				i, _ := strconv.Atoi(match[1])
				return i, true
			}

			if match := gccStyle.FindStringSubmatch(checkLine); match != nil {
				i, _ := strconv.Atoi(match[1])
				return i, true
			}
		}
	}

	return -1, false
}
