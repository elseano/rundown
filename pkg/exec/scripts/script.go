package scripts

import (
	"io/ioutil"
	"os"
	"strings"
)

type Script struct {
	AbsolutePath     string
	Contents         []byte
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

	tempFile.Write([]byte("#!/usr/bin/env "))
	tempFile.Write([]byte(s.Invocation))
	tempFile.Write([]byte("\n\n"))
	tempFile.Write(s.Contents)

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
