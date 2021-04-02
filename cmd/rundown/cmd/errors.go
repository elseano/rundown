package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/logrusorgru/aurora"
)

var (
	ErrorInternal   = errors.New("Internal rundown error")
	ErrorInvocation = errors.New("Invalid section provided")
	ErrorArg        = errors.New("Invalid section options provided")
	ErrorStopFail   = errors.New("Rundown file signalled failure")
	ErrorScript     = errors.New("Codeblock failed")
)

func handleError(dest io.Writer, err error, cb func()) error {
	defer func() {
		if cb != nil {
			fmt.Fprintf(dest, "\n")
			cb()
			fmt.Fprintf(dest, "\n")
		}
	}()

	if err == nil {
		return nil
	}

	if stopError, ok := err.(*rundown.StopError); ok {
		if stopError.Result.IsError {

			if stopError.Result.Kind == "Error" {

				fmt.Fprintf(dest, "\n%s - %s", aurora.Red("Error"), stopError.Result.Message)
				fmt.Fprintf(dest, " in:\n\n")

				for i, line := range strings.Split(strings.TrimSpace(stopError.Result.Source), "\n") {
					if i == stopError.Result.FocusLine-1 {
						fmt.Fprintf(dest, aurora.Faint("%3d:").String()+" %s\n", i+1, aurora.Red(line))
					} else {
						fmt.Fprintf(dest, aurora.Faint("%3d:").String()+" %s\n", i+1, line)
					}
				}

				fmt.Fprintf(dest, "\n")

				fmt.Fprintf(dest, stopError.Result.Output)
				return ErrorScript

			} else if stopError.Result.Message != "" {
				fmt.Fprintf(dest, "\n%s - %s", aurora.Red("Failure"), stopError.Result.Message)
				fmt.Fprintf(dest, "\n\n")
			} else {
				fmt.Fprintf(dest, "\n%s\n\n", aurora.Red("Failure"))
			}

			return ErrorStopFail
		}

		return nil // Stop requested.
	}

	fmt.Fprintf(dest, "\n%s: %s\n\n", aurora.Red("Error"), err)

	if errors.Is(err, rundown.InvocationError) {
		return ErrorInvocation
	}

	return ErrorInternal
}
