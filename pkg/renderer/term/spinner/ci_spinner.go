package spinner

import (
	"fmt"
	"io"
	"time"

	"github.com/logrusorgru/aurora"
)

type CISpinner struct {
	out              io.Writer
	currentHeading   string
	stampedHeading   string
	substep          string
	startedAt        time.Time
	substepStartedAt time.Time
	colors           aurora.Aurora
}

func NewCISpinner(out io.Writer, colors aurora.Aurora) *CISpinner {
	return &CISpinner{out: out, colors: colors}
}

func (s *CISpinner) Active() bool {
	return false
}

func (s *CISpinner) Start() {
	if s.startedAt.IsZero() {
		s.startedAt = time.Now()
	}
}

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
		s.out.Write([]byte(fmt.Sprintf("  %s %s %s\n", indicator, s.substep, buildTimeString(s.substepStartedAt))))
	}

	s.out.Write([]byte(fmt.Sprintf("%s %s %s\n", indicator, s.CurrentHeading(), buildTimeString(s.startedAt))))
	s.startedAt = time.Time{}
}

func (s *CISpinner) Success(message string) {
	s.closeSpinner(s.colors.Green(TICK).String())
}

func (s *CISpinner) Error(message string) {
	s.closeSpinner(s.colors.Red(CROSS).String())
}

func (s *CISpinner) Skip() {
	s.closeSpinner(s.colors.Yellow(SKIP).String())
}

func (s *CISpinner) SetMessage(message string) {
	s.currentHeading = message
}

func buildTimeString(t time.Time) string {
	return "(" + time.Since(t).Round(time.Millisecond).String() + ")"
}

func (s *CISpinner) NewStep(message string) {
	if s.substep != "" {
		s.out.Write([]byte(fmt.Sprintf("  %s %s %s\n", s.colors.Green(TICK), s.substep, s.colors.Faint(buildTimeString(s.substepStartedAt)))))
	} else {
		s.out.Write([]byte(fmt.Sprintf("%s %s\n", s.colors.Faint(DASH), s.CurrentHeading())))
	}

	s.substep = message
	s.substepStartedAt = time.Now()
}

func (s *CISpinner) HideAndExecute(f func()) {

}

func (s *CISpinner) CurrentHeading() string {
	return s.currentHeading
}
