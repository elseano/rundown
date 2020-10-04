package util

import (
	"regexp"
	"strings"
)

var colorMarker = regexp.MustCompile("\x1b\\[([0-9\\;]*[A-Za-z])")
var linkMarker = regexp.MustCompile("\x1b\\]8;;(.*?)\x1b\\\\(.*?)\x1b\\]8;;\x1b\\\\")

func RemoveColors(input string) string {
	decolored := colorMarker.ReplaceAllString(input, "")
	return linkMarker.ReplaceAllString(decolored, "$1|$2")
}

var returnsMatch = regexp.MustCompile("(^|\n)?.*?\r(.*?)(\n|$)")

func CollapseReturns(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")

	for {
		if !returnsMatch.MatchString(input) {
			break
		}

		input = returnsMatch.ReplaceAllString(input, "$1$2$3")
	}

	return input
}
