package util

import (
	"regexp"
	"strings"
)

var VariableDetection = regexp.MustCompile(`(?i)\$([a-z0-9_]+)|\${([a-z0-9_]+)(([^a-z0-9_]+)(.*?))?}`)

// Performs environment substitution on the source string.
// Supports $VAR, or ${VAR} optionally with the -, :-, +, :+ modifiers.
// Ignores other (or invalid) modifiers, substituting as if they weren't there.
func SubEnv(environment map[string]string, source string) string {
	for {
		matches := VariableDetection.FindStringSubmatch(source)

		if matches == nil {
			break
		}

		switch {
		case matches[2] == "":
			source = strings.ReplaceAll(source, matches[0], environment[matches[1]])
		case matches[2] != "":
			switch {
			case matches[4] == "-", matches[4] == ":-":
				if env := environment[matches[2]]; env == "" {
					source = strings.ReplaceAll(source, matches[0], matches[5])
				} else {
					source = strings.ReplaceAll(source, matches[0], env)
				}
			case matches[4] == "+", matches[4] == ":+":
				if env := environment[matches[2]]; env != "" {
					source = strings.ReplaceAll(source, matches[0], matches[5])
				} else {
					source = strings.ReplaceAll(source, matches[0], "")
				}

			default:
				source = strings.ReplaceAll(source, matches[0], environment[matches[2]])
			}
		}
	}

	return source
}
