package cmd

var (
	flagCodes     bool
	flagDebug     bool
	flagAsk       bool
	flagAskRepeat bool
	flagDefault   string
	flagCols      int

	argFilename   string
	argShortcodes = []string{}
)

var GitCommit string
var Version string
