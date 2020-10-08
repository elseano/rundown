package rundown

type ExecutionResult struct {
	Message   string
	Kind      string
	Source    string
	Output    string
	IsError   bool
	FocusLine int
}

var (
	SuccessfulExecution = ExecutionResult{Kind: "Success", IsError: false, FocusLine: -1}
	SkipToNextHeading   = ExecutionResult{Kind: "Skip", IsError: false, FocusLine: -1}
	StopFailResult      = ExecutionResult{Kind: "Stop", IsError: true, FocusLine: -1}
	StopOkResult        = ExecutionResult{Kind: "Stop", IsError: false, FocusLine: -1}
)

type StopError struct {
	Result ExecutionResult
}

func (e *StopError) Error() string {
	return e.Result.Message
}
