package errors

// Raised after running an execution block, indicates we want to skip to the next same-level heading
type JumpToNextHeadingError struct{}

func (JumpToNextHeadingError) Error() string { return "E_JUMP_NEXT_HEADING" }

type ExecutionError struct {
	Output []byte
	Err    error
}

func (ExecutionError) Error() string { return "E_SCRIPT_ERROR" }
