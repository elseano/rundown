package main

import (
	"github.com/elseano/rundown/cmd/rdv/cmd"
)

var GitCommit string
var Version string

func main() {
	cmd.Execute(Version, GitCommit)
}
