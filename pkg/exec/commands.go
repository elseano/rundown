package exec

import (
	"encoding/base64"
	"fmt"
	"regexp"

	"github.com/elseano/rundown/pkg/exec/scripts"
)

type RundownCommandHandler interface {
	SetSpinnerTitle(title string)
	SetEnvironmentVariable(name string, value string)
}

func ChangeCommentsToSpinnerCommands(sourceLang string, source []byte) []byte {
	switch sourceLang {
	case "bash", "sh", "fish":
	default:
		return source
	}

	var commentDetector = regexp.MustCompile(`(?m)^(\s*)#+\>\s+(.*)$`)

	return commentDetector.ReplaceAllFunc(source, func(match []byte) []byte {
		submatches := commentDetector.FindAllSubmatch(match, 1)

		// Base64 encode
		result := base64.StdEncoding.EncodeToString(submatches[0][2])
		indent := submatches[0][1]

		return []byte(fmt.Sprintf("%secho -n -e \"\x1b]R;SETSPINNER %s\x9c\"", indent, result))
	})
}

func AddEnvironmentCapture(sourceLang string, script *scripts.Script, captures []string) {
	switch sourceLang {
	case "bash", "sh", "fish":
	default:
		return
	}

	for _, envName := range captures {
		script.AppendCommand(fmt.Sprintf("echo -n -e \"\x1b]R;SETENV %s=$%s\x9c\"", envName, envName))
	}
}
