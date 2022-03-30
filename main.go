package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/elseano/rundown/cmd/rundown/cmd"
	"github.com/elseano/rundown/pkg/errs"
	"github.com/elseano/rundown/pkg/renderer/term"
	"github.com/elseano/rundown/pkg/util"
	"github.com/logrusorgru/aurora"
	"github.com/mattn/go-isatty"
)

var GitCommit string
var Version string

func main() {
	devNull, _ := os.Create(os.DevNull)
	util.RedirectLogger(devNull)

	useColors := true

	if !isatty.IsTerminal(os.Stdout.Fd()) {
		_, isCI := os.LookupEnv("CI")
		useColors = isCI
	} else if _, ok := os.LookupEnv("NO_COLOR"); ok {
		useColors = false
	}

	term.Aurora = aurora.NewAurora(useColors)
	term.ColorsEnabled = useColors

	var executionError *errs.ExecutionError
	err := cmd.Execute("", "")

	if errors.As(err, &executionError) {
		os.Exit(executionError.ExitCode)
	}

	if errors.Is(err, errs.ErrStopFail) {
		os.Exit(1)
	}

	if err != nil {
		fmt.Printf("Error: %s\n", err.Error())
		os.Exit(1)
	}
}
