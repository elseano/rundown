package util

import (
	"os"
	"os/user"
	"strings"
)

func FileExists(name string) bool {
	if name == "" {
		return false
	}

	expandedPath := name
	if strings.HasPrefix(expandedPath, "~") {
		Debugf("Prefixed with ~\n")
		if u, err := user.Current(); err == nil {
			Debugf("Get home dir\n")
			expandedPath = strings.Replace(expandedPath, "~", u.HomeDir, 1)
		} else {
			Debugf("Cannot get home dir %#v\n", err)
		}
	}

	Debugf("Opening file %s\n", expandedPath)

	file, err := os.Open(expandedPath)

	Debugf("Err: %#v\n", err)

	if err != nil {
		return os.IsExist(err)
	}

	file.Close()

	Debugf("All good.\n\n")

	return true
}
