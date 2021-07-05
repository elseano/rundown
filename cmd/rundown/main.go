package main

import (
	"os"

	"github.com/elseano/rundown/cmd/rundown/cmd"
)

var GitCommit string
var Version string

func main() {
	switch cmd.Execute(Version, GitCommit) {
	case cmd.ErrorArg:
		os.Exit(128)
	case cmd.ErrorInvocation:
		os.Exit(127)
	case cmd.ErrorInternal:
		os.Exit(129)
	case cmd.ErrorStopFail:
		os.Exit(2)
	case cmd.ErrorScript:
		os.Exit(1)
	case nil:
		os.Exit(0)
	}

	os.Exit(3)

}
