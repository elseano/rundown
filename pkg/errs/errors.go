package errs

type ExecutionError struct {
	ExitCode int
}

func (e *ExecutionError) Error() string {
	return "Cannot continue"
}
