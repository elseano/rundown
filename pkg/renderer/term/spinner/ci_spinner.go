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
	substep        string
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
	if s.substep == "" && s.stampedHeading != s.currentHeading {
		s.out.Write([]byte(s.colors.Faint(fmt.Sprintf("↓ %s\r\n", s.currentHeading)).String()))
		s.stampedHeading = s.currentHeading
	} else if s.substep != "" && s.stampedHeading != s.substep {
		s.out.Write([]byte(s.colors.Faint(fmt.Sprintf("  ↓ %s\r\n", s.substep)).String()))
		s.stampedHeading = s.substep
	}
}

func (s *CISpinner) closeSpinner(indicator string) {
	if s.substep != "" {
		s.out.Write([]byte(fmt.Sprintf("  %s %s\n", indicator, s.substep)))
	}

	s.out.Write([]byte(fmt.Sprintf("%s %s\n", indicator, s.CurrentHeading())))

}

func (s *CISpinner) Success(message string) {
	s.closeSpinner(s.colors.Green(TICK).String())
}

func (s *CISpinner) Error(message string) {
	s.closeSpinner(s.colors.Red(CROSS).String())
}

func (s *CISpinner) Skip(message string) {
	s.closeSpinner(s.colors.Yellow(SKIP).String())
}

func (s *CISpinner) SetMessage(message string) {
	s.currentHeading = message
}

func (s *CISpinner) NewStep(message string) {
	if s.substep != "" {
		s.out.Write([]byte(fmt.Sprintf("  %s %s\n", s.colors.Green(TICK), s.substep)))
	} else {
		s.out.Write([]byte(fmt.Sprintf("%s %s\n", s.colors.Faint(DASH), s.CurrentHeading())))
	}

	s.substep = message
}

func (s *CISpinner) HideAndExecute(f func()) {

}

func (s *CISpinner) CurrentHeading() string {
	return s.currentHeading
}
