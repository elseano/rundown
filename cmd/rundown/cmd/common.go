package cmd

var (
	flagCodes       bool
	flagDebug       bool
	flagAsk         bool
	flagAskRepeat   bool
	flagDefault     string
	flagCols        int
	flagFilename    string
	flagCompletions string

	flagViewOnly  bool
	flagCheckOnly bool
	flagAst       bool

	rundownFile   string
	argShortcodes = []string{}
)
