package scripts

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Script struct {
	AbsolutePath     string
	Prefix           []byte
	Contents         []byte
	Suffix           []byte
	Invocation       string
	EnvReferenceName string
	Name             string
	ShellScript      bool
}

func (s *Script) Write() error {
	tempFile, err := ioutil.TempFile("", "rd-"+strings.ToLower(s.Name)+"-*")
	if err != nil {
		return err
	}

	invo, prefix, err := buildInvocation(s.Invocation)
	if err != nil {
		return err
	}

	if invo != "" {
		tempFile.Write([]byte("#!"))
		tempFile.Write([]byte(invo))
		tempFile.Write([]byte("\n\n"))
	}

	if prefix != "" {
		if s.Prefix == nil {
			s.Prefix = []byte(fmt.Sprintf("%s\n", prefix))
		} else {
			s.Prefix = append(s.Prefix, []byte(fmt.Sprintf("%s\n", prefix))...)
		}
	}
	if s.Prefix != nil {
		tempFile.Write(s.Prefix)
		tempFile.Write([]byte("\n"))
	}
	tempFile.Write(s.Contents)
	if s.Suffix != nil {
		tempFile.Write([]byte("\n"))
		tempFile.Write(s.Suffix)
	}

	defer tempFile.Close()
	s.AbsolutePath = tempFile.Name()

	if invo != "" {
		os.Chmod(s.AbsolutePath, 0700)
	}

	return nil
}

func isShellLike(via string) bool {
	switch via {
	case "bash":
		return true
	case "sh":
		return true
	case "fish":
		return true
	default:
		return false
	}
}

func buildInvocation(interpreter string) (string, string, error) {
	// If no interpreter, just save the file.
	if interpreter == "" {
		return "", "", nil
	}

	abs, err := exec.LookPath(interpreter)

	if err != nil {
		return "", "", err
	}

	switch interpreter {
	case "bash":
		return abs, "set -euo pipefail", nil
	case "sh":
		return abs, "set -euo pipefail", nil
	default:
		return abs, "", nil
	}
}
