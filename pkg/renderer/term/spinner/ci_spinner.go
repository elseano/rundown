package spinner

import (
	"fmt"
	"io"

	"github.com/logrusorgru/aurora"
)

type CISpinner struct {
	out            io.Writer
	currentHeading string
	stampedHeading string
	colors         aurora.Aurora
}

func NewCISpinner(out io.Writer, colors aurora.Aurora) *CISpinner {
	return &CISpinner{out: out, colors: colors}
}

func (s *CISpinner) Active() bool {
	return false
}

func (s *CISpinner) Start() {}

func (s *CISpinner) Stop() {}

func (s *CISpinner) StampShadow() {
	if s.stampedHeading != s.currentHeading {
		s.out.Write([]byte(s.colors.Faint(fmt.Sprintf("↓ %s\r\n", s.currentHeading)).String()))
		s.stampedHeading = s.currentHeading
	}
}

func (s *CISpinner) Success(message string) {
	if message != "" {
		message = s.colors.Faint(fmt.Sprintf("(%s)", message)).String()
	}

	s.out.Write([]byte(fmt.Sprintf("%s %s %s\n", s.colors.Green(TICK), s.CurrentHeading(), message)))
}

func (s *CISpinner) Error(message string) {
	if message != "" {
		message = s.colors.Faint(fmt.Sprintf("(%s)", message)).String()
	}

	s.out.Write([]byte(fmt.Sprintf("%s %s %s\n", s.colors.Red(CROSS), s.CurrentHeading(), message)))
}

func (s *CISpinner) Skip(message string) {
	if message != "" {
		message = s.colors.Faint(fmt.Sprintf("(%s)", message)).String()
	}

	s.out.Write([]byte(fmt.Sprintf("%s %s %s\n", s.colors.Yellow(SKIP), s.CurrentHeading(), message)))
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
