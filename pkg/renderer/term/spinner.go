package term

// spx "github.com/tj/go-spin"
// spx "github.com/briandowns/spinner"

type Spinner interface {
	Active() bool
	Start()
	Stop()
	Success(message string)
	Error(message string)
	Skip(message string)
	SetMessage(message string)
	NewStep(message string)
	HideAndExecute(f func())
	CurrentHeading() string
	StampShadow()
}
