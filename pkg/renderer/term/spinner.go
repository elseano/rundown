package term

import (
	"io"
	"strings"
	"time"

	"github.com/elseano/rundown/pkg/renderer/term/spinner"
	// spx "github.com/tj/go-spin"
	// spx "github.com/briandowns/spinner"
)

const (
	TICK  = "✔"
	CROSS = "✖"
	SKIP  = "≡"
)

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
}

type RundownSpinner struct {
	s       *spinner.ActualSpinner
	indent  int
	message string
	out     io.Writer
}

func NewSpinner(indent int, message string, out io.Writer) Spinner {

	s := spinner.NewActualSpinner(spinner.CharSets[21], 100*time.Millisecond, spinner.WithWriter(out), spinner.WithColor("fgHiCyan"))
	s.Suffix = " " + message
	s.Prefix = strings.Repeat("  ", indent)

	return &RundownSpinner{indent: indent, message: message, s: s, out: out}
}

func (s *RundownSpinner) Active() bool {
	return s.s.Active()
}

func (s *RundownSpinner) Spin() {
	s.Start()
}

func (s *RundownSpinner) Start() {
	s.s.Start()
}

func (s *RundownSpinner) Success(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + Aurora.Green(TICK).String() + " " + s.message + " (" + Aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Error(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + Aurora.Red(CROSS).String() + " " + s.message + " (" + Aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Skip(message string) {
	s.s.FinalMSG = strings.Repeat("  ", s.indent) + Aurora.Faint(SKIP).String() + " " + s.message + " (" + Aurora.Faint(message).String() + ")\r\n"
	s.Stop()
}

func (s *RundownSpinner) Stop() {
	s.s.Stop()
}

func (s *RundownSpinner) CurrentHeading() string {
	return s.message
}

func (s *RundownSpinner) NewStep(message string) {
	var wasActive = s.Active()
	s.Success("OK")

	sp := spinner.NewActualSpinner(spinner.CharSets[21], 100*time.Millisecond, spinner.WithWriter(s.out), spinner.WithColor("fgHiCyan"))
	sp.Suffix = " " + message
	sp.Prefix = strings.Repeat("  ", s.indent)

	s.message = message
	s.s = sp

	if wasActive {
		s.Start()
	}
}

func (s *RundownSpinner) SetMessage(message string) {
	s.message = message
	s.s.Suffix = " " + message //+ "\033[1A"

	if s.s.Active() {
		s.s.Repaint()
	}
}

func (s *RundownSpinner) HideAndExecute(f func()) {
	s.s.HideAndExecute(f)
}
