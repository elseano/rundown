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

	invo, err := buildInvocation(s.Invocation)
	if err != nil {
		return err
	}

	tempFile.Write([]byte("#!"))
	tempFile.Write([]byte(invo))
	tempFile.Write([]byte("\n\n"))
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

	os.Chmod(s.AbsolutePath, 0700)

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

func buildInvocation(interpreter string) (string, error) {
	abs, err := exec.LookPath(interpreter)

	if err != nil {
		return "", err
	}

	switch interpreter {
	case "bash":
		return fmt.Sprintf("%s -euo pipefail", abs), nil
	case "sh":
		return fmt.Sprintf("%s -euo pipefail", abs), nil
	default:
		return abs, nil
	}
}
