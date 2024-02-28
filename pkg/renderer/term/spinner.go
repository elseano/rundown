package term

import "io"

// spx "github.com/tj/go-spin"
// spx "github.com/briandowns/spinner"

type Spinner interface {
	Active() bool
	Start()
	Stop()
	Success(message string)
	Error(message string)
	Skip()
	SetMessage(message string)
	NewStep(message string)
	HideAndExecute(f func())
	CurrentHeading() string
	StampShadow()
}

var NewSpinnerFunc func(w io.Writer) Spinner = nil
