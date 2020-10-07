package rundown

type ExecutionResult struct {
	Message string
	Kind    string
	Source  string
	Output  string
	IsError bool
}

var (
	SuccessfulExecution = ExecutionResult{Kind: "Success", IsError: false}
	SkipToNextHeading   = ExecutionResult{Kind: "Skip", IsError: false}
	StopFailResult      = ExecutionResult{Kind: "Stop", IsError: true}
	StopOkResult        = ExecutionResult{Kind: "Stop", IsError: false}
)

type StopError struct {
	Result ExecutionResult
}

func (e *StopError) Error() string {
	return e.Result.Message
}
