package cmd

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/elseano/rundown/pkg/rundown"
	"github.com/logrusorgru/aurora"
)

const (
	ErrorCodeInternal   = 127 // Something failed with Rundown internally.
	ErrorInvocation     = 3   // Invalid shortcode or options provided.
	ErrorCodeUnexpected = 2   // A script failed and we weren't expecting it.
	ErrorCodeExpected   = 1   // Stop-Fail requested.
	ErrorCodeSuccess    = 0   // Stop-Ok or normal script termination.
)

func handleError(dest io.Writer, err error, cb func()) int {
	defer func() {
		if cb != nil {
			fmt.Fprintf(dest, "\n")
			cb()
			fmt.Fprintf(dest, "\n")
		}
	}()

	if err == nil {
		return ErrorCodeSuccess
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
				return ErrorCodeUnexpected

			} else if stopError.Result.Message != "" {
				fmt.Fprintf(dest, "\n%s - %s", aurora.Red("Failure"), stopError.Result.Message)
				fmt.Fprintf(dest, "\n\n")
			} else {
				fmt.Fprintf(dest, "\n%s\n\n", aurora.Red("Failure"))
			}

			return ErrorCodeExpected
		}

		return ErrorCodeSuccess // Stop requested.
	}

	fmt.Fprintf(dest, "\n%s: %s\n\n", aurora.Red("Error"), err)

	if errors.Is(err, rundown.InvocationError) {
		return ErrorInvocation
	}

	return ErrorCodeInternal
}
