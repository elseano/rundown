package util

import (
	"os"
	"os/user"
	"path"
	"path/filepath"
	"strings"
)

func MakeRelative(filename string) string {
	wd, err := os.Getwd()

	if err != nil {
		return ""
	}

	rel, err := filepath.Rel(wd, filename)
	if err != nil {
		return err.Error()
	}

	return rel
}

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

// Finds a file in the current directory or parents. Files are checked in the order they're provided.
func FindFile(files []string) string {
	dir, err := os.Getwd()

	if err != nil {
		// If we can't get the CWD, just search the current path.
		for _, filename := range files {
			if FileExists(filename) {
				return filename
			}
		}

		return ""
	}

	// Search the CWD and parent paths until the root path for RUNDOWN.md/README.md
	for {
		for _, fn := range files {
			path := path.Join(dir, fn)
			Debugf("Searching: %s\n", path)

			if FileExists(path) {
				return path
			}
		}

		if dir == "/" {
			// Reached the end
			return ""
		}

		dir = path.Dir(dir)
	}
}
