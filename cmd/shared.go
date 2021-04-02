package cmd

import "github.com/elseano/rundown/pkg/util"

func RundownFile(fileFlag string) string {
	if fileFlag != "" {
		return fileFlag
	}

	return util.FindFile([]string{"RUNDOWN.md", "README.md"})
}
