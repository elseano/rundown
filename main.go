package main

import (
	"os"

	"github.com/elseano/rundown/cmd/rundown/cmd"
	"github.com/elseano/rundown/pkg/util"
)

var GitCommit string
var Version string

func main() {
	devNull, _ := os.Create(os.DevNull)
	util.RedirectLogger(devNull)

	cmd.Execute("", "")
}
