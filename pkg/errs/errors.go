package errs

import "errors"

type ExecutionError struct {
	ExitCode int
}

func (e *ExecutionError) Error() string {
	return "Cannot continue"
}

var ErrStopOk = errors.New("StopOK")
var ErrStopFail = errors.New("StopFail")
