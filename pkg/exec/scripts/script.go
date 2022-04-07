package scripts

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/elseano/rundown/pkg/util"
)

type Script struct {
	tempFile         *os.File
	AbsolutePath     string
	Prefix           []byte
	Contents         []byte
	OriginalContents []byte
	Suffix           []byte
	BinaryPath       string
	CommandLine      string
	EnvReferenceName string
	Name             string
	ShellScript      bool
}

func NewScript(binary string, language string, contents []byte) (*Script, error) {
	tempFile, err := ioutil.TempFile("", "rd-*")
	if err != nil {
		return nil, err
	}

	binaryPath, prefix, err := buildInvocation(binary, language)

	if err != nil {
		return nil, err
	}

	var commandline string
	if strings.Contains(binaryPath, "$SCRIPT_FILE") {
		commandline = binaryPath
	} else {
		commandline = fmt.Sprintf("%s %s", binaryPath, tempFile.Name())
	}

	return &Script{OriginalContents: contents, CommandLine: commandline, BinaryPath: binaryPath, Contents: contents, tempFile: tempFile, AbsolutePath: tempFile.Name(), Prefix: []byte(prefix)}, nil
}

func (s *Script) MakeExecutable() {
	os.Chmod(s.AbsolutePath, 0700)
}

func (s *Script) AppendCommand(command string) {
	if s.Suffix == nil {
		s.Suffix = make([]byte, 0)
	}

	s.Suffix = append(s.Suffix, []byte(command)...)
	s.Suffix = append(s.Suffix, []byte("\n")...)
}

func (s *Script) Write() error {
	defer s.tempFile.Close()

	result := bytes.Buffer{}

	if len(s.Prefix) > 0 {
		result.Write(s.Prefix)
		result.Write([]byte("\n"))
	}

	result.Write(s.Contents)

	if s.Suffix != nil {
		result.Write([]byte("\n"))
		result.Write(s.Suffix)
	}

	util.Logger.Debug().Msgf("FINAL Script is: %s", result.String())
	_, err := s.tempFile.Write(result.Bytes())

	return err
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

func buildInvocation(binary string, language string) (string, string, error) {
	// If no interpreter, just save the file.
	if binary == "" {
		return "", "", nil
	}

	abs, err := exec.LookPath(binary)

	// Ignore not found errors
	if lpErr, ok := err.(*exec.Error); ok && lpErr.Err == exec.ErrNotFound {
		err = nil
		abs = binary
	}

	if err != nil {
		return "", "", err
	}

	switch language {
	case "bash":
		return abs, "set -euo pipefail", nil
	case "sh":
		// pipefail on Ubuntu's SH fails, so don't set it here.
		return abs, "set -eu", nil
	default:
		return abs, "", nil
	}
}
