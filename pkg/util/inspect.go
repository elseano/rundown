package util

import (
	"strings"
)

func InspectString(s string) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\033", "\\033")
	return s
}

func InspectList(l []string) string {
	var result []string

	for _, s := range l {
		result = append(result, InspectString(s))
	}

	return strings.Join(result, ", ")
}