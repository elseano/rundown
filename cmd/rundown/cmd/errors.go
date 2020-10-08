package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/logrusorgru/aurora"
)

func handleError(err error) {
	if err == nil {
		return
	}

	if stopError, ok := err.(*rundown.StopError); ok {
		if stopError.Result.IsError {
			fmt.Printf("\n\n%s - %s in:\n\n", aurora.Bold("Error"), stopError.Result.Message)
			for i, line := range strings.Split(strings.TrimSpace(stopError.Result.Source), "\n") {
				if i == stopError.Result.FocusLine-1 {
					fmt.Printf(aurora.Faint("%3d:").String()+" %s\n", i+1, aurora.Red(line))
				} else {
					fmt.Printf(aurora.Faint("%3d:").String()+" %s\n", i+1, line)
				}
			}

			fmt.Println()

			fmt.Println(stopError.Result.Output)
			os.Exit(127)
		}

		os.Exit(0) // Stop requested.
	}

	fmt.Printf("\n\n\n%s: %s (%T)\n", aurora.Bold("Error"), err.Error(), err)
	os.Exit(1)
}
