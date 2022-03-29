package spinner

import (
	"fmt"
	"io"
)

type CISpinner struct {
	out            io.Writer
	currentHeading string
}

func NewCISpinner(out io.Writer) *CISpinner {
	return &CISpinner{out: out}
}

func (s *CISpinner) Active() bool {
	return false
}

func (s *CISpinner) Start() {}

func (s *CISpinner) Stop() {}

func (s *CISpinner) StampShadow() {
}

func (s *CISpinner) Success(message string) {
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", TICK, s.CurrentHeading(), message)))
}

func (s *CISpinner) Error(message string) {
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", CROSS, s.CurrentHeading(), message)))
}

func (s *CISpinner) Skip(message string) {
	s.out.Write([]byte(fmt.Sprintf("  %s %s (%s)", SKIP, s.CurrentHeading(), message)))
}

func (s *CISpinner) SetMessage(message string) {
	s.currentHeading = message
}

func (s *CISpinner) NewStep(message string) {
	s.SetMessage(message)
}

func (s *CISpinner) HideAndExecute(f func()) {

}

func (s *CISpinner) CurrentHeading() string {
	return s.currentHeading
}
