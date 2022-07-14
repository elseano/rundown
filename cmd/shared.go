package cmd

import (
	"path/filepath"

	"github.com/elseano/rundown/pkg/util"
)

func RundownFile(fileFlag string) string {
	util.Debugf("File flag is %s\n", fileFlag)

	if fileFlag != "" {
		abs, err := filepath.Abs(fileFlag)
		if err != nil {
			return fileFlag
		}

		return abs
	}

	return util.FindFile([]string{"RUNDOWN.md", "rundown.md"})
}
