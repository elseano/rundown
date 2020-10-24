package main

import (
	"github.com/elseano/rundown/cmd/rundown/cmd"
)

var GitCommit string
var Version string

func main() {
	cmd.Execute(Version, GitCommit)
}
