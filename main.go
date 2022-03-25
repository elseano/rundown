package main

import (
	"os"

	"github.com/elseano/rundown/cmd/rundown/cmd"
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
		useColors = false
	} else if _, ok := os.LookupEnv("NO_COLOR"); ok {
		useColors = false
	}

	term.Aurora = aurora.NewAurora(useColors)
	term.ColorsEnabled = useColors

	cmd.Execute("", "")
}
